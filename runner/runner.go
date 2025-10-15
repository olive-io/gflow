/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/clientgo"
	"github.com/olive-io/gflow/pkg/version"
)

type Runner struct {
	lg  *zap.Logger
	cfg *Config

	name string

	tr *atomic.Pointer[types.Runner]

	endpoints map[string]*types.Endpoint

	taskDefines map[string]Task

	pmu   sync.RWMutex
	pools map[string]Task
}

func New(name string, cfg *Config) (*Runner, error) {

	lg := cfg.Logger()
	cpuTotal := uint64(0)
	cpus, err := cpu.Counts(false)
	if err != nil {
		return nil, fmt.Errorf("read system cpu: %w", err)
	}
	cpuInfos, _ := cpu.Info()
	if len(cpuInfos) > 0 {
		cpuTotal = uint64(cpus) * uint64(cpuInfos[0].Mhz)
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("read system memory: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("read system hostname: %w", err)
	}

	tr := &types.Runner{
		Uid:         cfg.ID,
		Version:     version.GitTag,
		HeartbeatMs: cfg.HeartBeatInterval.Milliseconds(),
		Hostname:    hostname,
		Metadata:    cfg.Metadata,
		Features:    make(map[string]string),
		Cpu:         cpuTotal,
		Memory:      vm.Total,
	}
	trPtr := atomic.Pointer[types.Runner]{}
	trPtr.Store(tr)

	runner := &Runner{
		lg:          lg,
		cfg:         cfg,
		name:        name,
		tr:          &trPtr,
		endpoints:   make(map[string]*types.Endpoint),
		taskDefines: make(map[string]Task),
		pools:       make(map[string]Task),
	}

	return runner, nil
}

func (r *Runner) Start(ctx context.Context) error {
	lg := r.lg

	runner := r.tr.Load()
	for _, targetURL := range r.cfg.TargetURLs() {
		ccfg := clientgo.NewConfig(targetURL)
		if r.cfg.CertFile != "" && r.cfg.KeyFile != "" {
			ccfg.TLS = &clientgo.ConfigTLS{
				CertFile: r.cfg.CertFile,
				KeyFile:  r.cfg.KeyFile,
				CaFile:   r.cfg.CaFile,
			}
		}

		gfc, err := clientgo.NewClient(ccfg)
		if err != nil {
			return fmt.Errorf("connect to %s: %w", targetURL, err)
		}

		runnerUID := runner.Uid
		endpoints := r.ListEndpoints()
		err = gfc.AddEndpoints(ctx, endpoints, runnerUID)
		if err != nil {
			return fmt.Errorf("add endpoints: %w", err)
		}

		dispatcher, err := gfc.NewDispatcher(ctx, lg, runner)
		if err != nil {
			lg.Fatal("create dispatcher", zap.Error(err))
		}

		id := dispatcher.ID()
		lg.Info("start dispatcher",
			zap.String("id", id),
			zap.String("target", targetURL))

		go func() {
			for {
				event, derr := dispatcher.Next()
				if derr != nil {
					if derr == io.EOF {
						continue
					}
					lg.Error("dispatch error",
						zap.String("id", id),
						zap.Error(derr))
					return
				}

				if event.Err != nil {
					lg.Error("dispatch next error",
						zap.String("id", id),
						zap.Error(event.Err))
					continue
				}

				if callMsg := event.Call; callMsg != nil {
					resp := r.Handle(ctx, callMsg)
					if len(resp.Error) != 0 {
						lg.Error("dispatch call error",
							zap.String("id", id),
							zap.String("error", resp.Error))
					}
					if err = dispatcher.CallReply(resp); err != nil {
						lg.Error("dispatch reply",
							zap.String("id", id),
							zap.Error(err))
					}
				}
			}
		}()

		go func() {
			interval := r.cfg.HeartBeatInterval
			if interval == 0 {
				interval = time.Second * 30
			}

			timer := time.NewTimer(interval)
			defer timer.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					stat := r.generateRunnerStat()
					if err = dispatcher.Heartbeat(stat); err != nil {
						lg.Error("dispatch heartbeat",
							zap.String("id", id),
							zap.Error(err))
					}
				}
			}
		}()
	}

	<-ctx.Done()

	lg.Sugar().Infof("shutdown %s", r.name)
	return nil
}

// generateRunnerStat returns latest *types.RunnerStat
func (r *Runner) generateRunnerStat() *types.RunnerStat {
	tr := r.tr.Load()
	rs := &types.RunnerStat{
		Uid:           tr.Uid,
		Timestamp:     time.Now().UnixMilli(),
		Steps:         uint64(stepCounter.Get()),
		CommitCount:   uint64(stepCommitCounter.Get()),
		RollbackCount: uint64(stepRollbackCounter.Get()),
		DestroyCount:  uint64(stepDestroyCounter.Get()),
	}
	interval := time.Millisecond * 300
	percents, _ := cpu.Percent(interval, false)
	if len(percents) > 0 {
		rs.CpuUsed = percents[0] * float64(tr.Cpu) / 100
	}

	vm, err := mem.VirtualMemory()
	if err == nil {
		rs.MemoryUsed = vm.UsedPercent * float64(tr.Memory)
	}

	return rs
}
