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

package scheduler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
)

var (
	ErrSchedulerStopped = errors.New("scheduler has been stopped")

	ErrIsClosed = errors.New("watch channel closed")
)

type WatchMessage struct {
	Err      error
	Process  *types.Process
	FlowNode *types.FlowNode
}

type WatchChan struct {
	ctx context.Context

	name string
	lg   *zap.Logger

	ch   chan *WatchMessage
	done chan struct{}

	closed <-chan struct{}
}

func newWatchChan(ctx context.Context, name string, lg *zap.Logger, closed <-chan struct{}) *WatchChan {
	w := &WatchChan{
		ctx:    ctx,
		name:   name,
		lg:     lg,
		ch:     make(chan *WatchMessage, 3),
		done:   make(chan struct{}, 1),
		closed: closed,
	}
	return w
}

func (w *WatchChan) send(msg *WatchMessage) {
	select {
	case <-w.ctx.Done():
	case <-w.done:
	case <-w.closed:
	case w.ch <- msg:
	}
}

func (w *WatchChan) Next() *WatchMessage {
	select {
	case <-w.ctx.Done():
		return &WatchMessage{Err: w.ctx.Err()}
	case <-w.done:
		return &WatchMessage{Err: ErrIsClosed}
	case <-w.closed:
		return &WatchMessage{Err: ErrSchedulerStopped}
	case msg, ok := <-w.ch:
		if !ok {
			return &WatchMessage{Err: ErrIsClosed}
		}
		return msg
	}
}

func (w *WatchChan) IsClosed() bool {
	select {
	case <-w.ctx.Done():
		return true
	case <-w.done:
		return true
	default:
		return false
	}
}

func (w *WatchChan) Close() {
	select {
	case <-w.done:
		return
	default:
		w.lg.Sugar().Infof("close watch [%s]", w.name)
		close(w.done)
	}
}
