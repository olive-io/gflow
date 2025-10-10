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

package dispatch

import (
	"context"
	"errors"
	"fmt"

	"github.com/olive-io/gflow/api/types"
)

var (
	ErrNotFound = errors.New("target not found")
	ErrNoTarget = errors.New("no target")
)

type CallOptions struct {
	UID string
}

type CallOption func(*CallOptions)

func WithUID(uid string) CallOption {
	return func(o *CallOptions) {
		o.UID = uid
	}
}

func (d *Dispatcher) CallTask(ctx context.Context, req *types.CallTaskRequest, opts ...CallOption) (*types.CallTaskResponse, error) {
	var options CallOptions
	for _, opt := range opts {
		opt(&options)
	}

	target, err := d.selectOne(options)
	if err != nil {
		return nil, err
	}

	newId := allocId()
	outCh := make(chan *event, 1)
	d.emu.Lock()
	d.events[newId] = outCh
	d.emu.Unlock()

	defer func() {
		d.emu.Lock()
		delete(d.events, newId)
		d.emu.Unlock()

		freeId(newId)
	}()

	req.SeqId = newId
	if serr := target.Send(&Request{CallTask: req}); serr != nil {
		return nil, serr
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-outCh:
		return resp.callTask, nil
	}
}

func (d *Dispatcher) selectOne(options CallOptions) (*pipe, error) {
	d.emu.RLock()
	defer d.emu.RUnlock()

	if uid := options.UID; uid != "" {
		target, ok := d.pipes[uid]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, uid)
		}
		return target, nil
	}

	for _, item := range d.pipes {
		return item, nil
	}

	return nil, ErrNoTarget
}
