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

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/clientgo"
	"github.com/olive-io/gflow/pkg/signalutil"
)

func main() {
	lg, _ := zap.NewDevelopment()
	cfg := clientgo.NewConfig("localhost:6550")
	gfc, err := clientgo.NewClient(cfg)
	if err != nil {
		lg.Fatal("create gflow client", zap.Error(err))
	}

	ctx := signalutil.SetupSignalContext(context.Background())
	hostname, _ := os.Hostname()
	runner := &types.Runner{
		Uid:         "node1",
		Name:        "gflow-node1",
		ListenUrl:   "",
		Version:     "v0.0.1",
		HeartbeatMs: 30000,
		Hostname:    hostname,
		Metadata:    map[string]string{},
		Features:    map[string]string{},
		Transport:   types.Runner_GRPCStream,
		Cpu:         2,
		Memory:      4096 * 1024 * 1024,
	}
	dispatcher, err := gfc.NewDispatcher(ctx, lg, runner)
	if err != nil {
		lg.Fatal("create dispatcher", zap.Error(err))
	}
	defer dispatcher.Close()

	go func() {
		for {
			next, err := dispatcher.Next()
			if err != nil {
				if err == io.EOF {
					continue
				}
				lg.Error("dispatch error", zap.Error(err))
				return
			}

			if next.Err != nil {
				lg.Error("dispatch next error", zap.Error(next.Err))
				continue
			}

			if callMsg := next.Call; callMsg != nil {
				resp, err := call(ctx, lg, callMsg)
				if err != nil {
					resp = &types.CallTaskResponse{SeqId: callMsg.SeqId, Error: err.Error()}
					lg.Error("dispatch call error", zap.Error(err))
				}
				_ = dispatcher.CallReply(resp)
			}
		}
	}()

	go func() {
		timer := time.NewTimer(time.Second * 15)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				stat := &types.RunnerStat{
					Uid:           runner.Uid,
					CpuUsed:       0,
					MemoryUsed:    0,
					State:         "",
					Error:         "",
					Timestamp:     0,
					Steps:         0,
					CommitCount:   0,
					RollbackCount: 0,
					DestroyCount:  0,
				}
				_ = dispatcher.Heartbeat(stat)
			}
		}
	}()

	select {
	case <-ctx.Done():

	}

	lg.Info("shutdown gracefully")
}

func call(ctx context.Context, lg *zap.Logger, req *types.CallTaskRequest) (*types.CallTaskResponse, error) {
	results := make(map[string]*types.Value)
	results["a"] = types.NewValue("bb")

	resp := &types.CallTaskResponse{
		SeqId:   req.SeqId,
		Results: results,
	}

	lg.Info("call response", zap.Any("request", req))

	return resp, fmt.Errorf("internal error")
}
