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
	"fmt"
	"sync"

	pb "github.com/olive-io/gflow/api/rpc"
)

type Request struct {
	CallTask *pb.CallTaskRequest
}

type Response struct {
	CallTask *pb.CallTaskResponse
}

type pipe struct {
	ctx context.Context

	uid string

	out chan *Request
	pch chan string
}

func (p *pipe) Send(req *Request) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.out <- req:
		return nil
	}
}

func (p *pipe) process() {
	defer close(p.out)
	for {
		select {
		case <-p.ctx.Done():
			p.pch <- p.uid
			return
		}
	}
}

type Dispatcher struct {
	pmu   sync.RWMutex
	pipes map[string]*pipe

	emu    sync.RWMutex
	events map[uint64]chan *pb.CallTaskResponse

	pch chan string
}

func NewDispatcher() *Dispatcher {
	dispatcher := &Dispatcher{
		pipes:  make(map[string]*pipe),
		events: make(map[uint64]chan *pb.CallTaskResponse),
		pch:    make(chan string, 10),
	}
	return dispatcher
}

func (d *Dispatcher) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case uid := <-d.pch:
			d.pmu.Lock()
			delete(d.pipes, uid)
			d.pmu.Unlock()
		}
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	go d.process(ctx)
}

// Register adds new pipe
func (d *Dispatcher) Register(ctx context.Context, uid string) (chan *Request, error) {
	d.pmu.RLock()
	_, ok := d.pipes[uid]
	d.pmu.RUnlock()

	if ok {
		return nil, fmt.Errorf("pipe %s already registered", uid)
	}

	ch := make(chan *Request, 1)
	p := &pipe{
		ctx: ctx,
		uid: uid,
		out: ch,
		pch: d.pch,
	}
	go p.process()

	d.pmu.Lock()
	d.pipes[uid] = p
	d.pmu.Unlock()

	return ch, nil
}

func (d *Dispatcher) CallTask(ctx context.Context, req *pb.CallTaskRequest) (*pb.CallTaskResponse, error) {

	var target *pipe
	d.pmu.RLock()
	for _, item := range d.pipes {
		target = item
	}
	d.pmu.RUnlock()

	if target == nil {
		return nil, fmt.Errorf("target not found")
	}

	newId := allocId()
	outCh := make(chan *pb.CallTaskResponse, 1)
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

	if err := target.Send(&Request{CallTask: req}); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-outCh:
		return resp, nil
	}
}

func (d *Dispatcher) Reply(resp *Response) {
	if msg := resp.CallTask; msg != nil {
		id := msg.SeqId
		d.emu.Lock()
		ch, ok := d.events[id]
		if ok {
			ch <- msg
		}
		d.emu.Unlock()
	}
}
