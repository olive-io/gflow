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

package clientgo

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/olive-io/gflow/api/rpc"
	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/pkg/logutil"
)

type Event struct {
	Call *pb.CallTaskRequest
	Err  error
}

type Dispatcher struct {
	ctx context.Context

	lg *zap.Logger

	runner *types.Runner

	systemClient pb.SystemRPCClient

	stream pb.SystemRPC_RunnerDispatchClient

	callOptions []grpc.CallOption

	connected *atomic.Bool

	ech chan *Event

	done chan struct{}
}

func (c *Client) NewDispatcher(ctx context.Context, lg *zap.Logger, runner *types.Runner) (*Dispatcher, error) {
	opts := c.buildCallOptions()
	if lg == nil {
		lcfg := logutil.NewLogConfig()
		lcfg.SetupGlobalLoggers()
		lg = lcfg.GetLogger()
	}

	connected := &atomic.Bool{}
	connected.Store(false)
	dispatcher := &Dispatcher{
		ctx:          ctx,
		lg:           lg,
		runner:       runner,
		systemClient: c.systemClient,
		callOptions:  opts,
		connected:    connected,
		ech:          make(chan *Event, 10),
		done:         c.done,
	}
	_, err := dispatcher.connect(ctx)
	if err != nil {
		return nil, err
	}

	go dispatcher.process()
	return dispatcher, nil
}

func (d *Dispatcher) Context() context.Context {
	return d.ctx
}

func (d *Dispatcher) connect(ctx context.Context) (*pb.HandshakeResponse, error) {
	d.lg.Info("connecting to gflow dispatch")
	stream, err := d.systemClient.RunnerDispatch(ctx, d.callOptions...)
	if err != nil {
		return nil, err
	}
	d.stream = stream

	msg := &pb.RunnerDispatchRequest{
		Handshake: &pb.HandshakeRequest{Runner: d.runner},
	}
	err = stream.Send(msg)
	if err != nil {
		return nil, parseErr(err)
	}

	out, err := stream.Recv()
	if err != nil {
		return nil, parseErr(err)
	}
	rsp := out.Handshake

	d.lg.Info("connect to gflow dispatch succeeded")
	d.connected.Store(true)
	return rsp, nil
}

func (d *Dispatcher) Heartbeat(stat *types.RunnerStat) error {
	msg := &pb.RunnerDispatchRequest{
		Heartbeat: &pb.HeartBeatRequest{Stat: stat},
	}

	err := d.stream.Send(msg)
	return parseErr(err)
}

func (d *Dispatcher) Next() (*Event, error) {
	select {
	case <-d.done:
		return nil, io.EOF
	case <-d.ctx.Done():
		return nil, d.ctx.Err()
	case e := <-d.ech:
		return e, nil
	}
}

func (d *Dispatcher) process() {
	ctx := d.Context()
	attempts := 0
	var interval time.Duration

	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		interval = retryInterval(attempts)
		timer.Reset(interval)

		attempts += 1

		select {
		case <-d.done:
			d.lg.Info("dispatch disconnected")
			return
		case <-timer.C:
		}

		if !d.connected.Load() {
			d.lg.Info("reconnecting to gflow dispatch")
			_, err := d.connect(ctx)
			if err != nil {
				if isUnavailable(err) {
					d.lg.Error("gflow dispatch is unavailable")
				} else {
					d.lg.Error("connects to gflow dispatch error", zap.Error(err))
					event := &Event{Err: err}
					d.ech <- event
				}
				d.connected.Store(false)
				continue
			} else {
				// 重连成功，重置连接间隔
				attempts = 0
			}
		}

		d.lg.Info("receives dispatch message")
	LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			case <-d.done:
				return
			default:
			}

			rsp, err := d.stream.Recv()
			if err != nil {
				if isUnavailable(err) {
					d.lg.Error("gflow dispatch is unavailable")
				} else {
					d.lg.Error("connect to gflow dispatch error", zap.Error(err))
					event := &Event{Err: err}
					d.ech <- event
				}
				d.connected.Store(false)
				break LOOP
			}

			event := &Event{Call: rsp.CallTask}
			d.ech <- event
		}
	}
}
