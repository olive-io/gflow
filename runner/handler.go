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
	"reflect"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
)

type Task interface {
	Commit(ctx context.Context, request any) (any, error)
	Rollback(ctx context.Context) error
	Destroy(ctx context.Context) error
	String() string
}

var _ Task = (*taskImpl)(nil)

type taskImpl struct {
	Task

	opt *Options
}

func (impl *taskImpl) Commit(ctx context.Context, in any) (any, error) {
	request := in.(*TaskRequest)

	argsType := reflect.TypeOf(reflect.TypeOf(impl.opt.Request).Elem())
	arg := reflect.New(argsType).Interface()
	if err := request.InjectFor(arg); err != nil {
		return nil, err
	}

	call := func(ctx context.Context, req any) (resp any, err error) {
		defer func() { err = doRecover() }()

		resp, err = impl.Task.Commit(ctx, arg)
		return
	}

	return call(ctx, request)
}

func (impl *taskImpl) Rollback(ctx context.Context) error {
	call := func(ctx context.Context) (err error) {
		defer func() { err = doRecover() }()

		err = impl.Task.Rollback(ctx)
		return
	}

	return call(ctx)
}

func (impl *taskImpl) Destroy(ctx context.Context) error {
	call := func(ctx context.Context) (err error) {
		defer func() { err = doRecover() }()

		err = impl.Task.Destroy(ctx)
		return
	}

	return call(ctx)
}

var _ Task = (*taskForFunc)(nil)

type taskForFunc struct {
	name      string
	methodPtr reflect.Value
	args      []reflect.Type

	ctxIsFirst  bool
	containsReq bool
}

func (fn *taskForFunc) Commit(ctx context.Context, arg any) (any, error) {
	request := arg.(*TaskRequest)

	method := fn.methodPtr

	inputs := make([]reflect.Value, 0)
	if fn.ctxIsFirst {
		inputs = append(inputs, reflect.ValueOf(ctx))
	}
	if fn.containsReq {
		in := fn.args[0]
		target := reflect.New(in.Elem()).Interface()
		if err := request.InjectFor(target); err != nil {
			return nil, err
		}
		inputs = append(inputs, reflect.ValueOf(target))
	} else {
		for i, in := range fn.args {
			key := fmt.Sprintf("p%d", i)
			tv, ok := request.Properties[key]
			if !ok {
				return nil, fmt.Errorf("missing '%s' property", key)
			}
			rv := reflect.New(in.Elem())
			if err := InjectFromTypesValue(rv, tv); err != nil {
				return nil, err
			}
			inputs = append(inputs, rv)
		}
	}

	call := func(in []reflect.Value) (outs []reflect.Value, err error) {
		defer func() { err = doRecover() }()
		outs = method.Call(in)
		if len(outs) > 0 {
			lastIdx := len(outs) - 1
			lastOut := outs[lastIdx]
			var ok bool
			err, ok = lastOut.Interface().(error)
			if ok {
				outs = outs[:lastIdx]
			}
		}
		return
	}

	outputs, cerr := call(inputs)
	if cerr != nil {
		return nil, cerr
	}

	resp := &TaskResponse{}
	switch len(outputs) {
	case 0:
	case 1:
		out := outputs[0]
		var ok bool
		resp, ok = out.Interface().(*TaskResponse)
		if ok {
			return resp, nil
		}

		if out.Kind() == reflect.Pointer {
			out = out.Elem()
		}

		if out.Kind() == reflect.Struct {
			rv := out.Type()
			for i := 0; i < rv.NumField(); i++ {
				ft := rv.Field(i)
				fv := out.Field(i)

				tag, found := ft.Tag.Lookup("json")
				if found {
					name := strings.Split(tag, ",")[0]
					vv := fv.Interface()
					tv := types.NewValue(vv)

					resp.Results[name] = tv

					continue
				}

				tag, found = ft.Tag.Lookup(DefaultTag)
				if !found {
					continue
				}

				parts := strings.Split(tag, ";")
				pairs := map[string]string{}
				for _, part := range parts {
					kv := strings.Split(part, ":")
					if len(kv) != 2 {
						continue
					}
					pairs[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])

					key, ok := pairs[DataObjectTag]
					if ok {
						vv := fv.Interface()
						tv := types.NewValue(vv)
						resp.Results[key] = tv
					}
				}
			}

		} else {
			tv := fromReflectValue(out)
			resp.Results["r0"] = tv
		}
	default:
		p := 0
		for _, out := range outputs {
			key := fmt.Sprintf("r%d", p)
			tv := types.NewValue(out.Interface())
			resp.Results[key] = tv
			p += 1
		}
	}

	return resp, nil
}

func (fn *taskForFunc) Rollback(ctx context.Context) error { return nil }

func (fn *taskForFunc) Destroy(ctx context.Context) error { return nil }

func (fn *taskForFunc) String() string { return fn.name }

type Options struct {
	Name     string
	Request  any
	Response any
}

type Option func(*Options)

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithRequest(request any) Option {
	return func(o *Options) {
		o.Request = request
	}
}

func WithResponse(response any) Option {
	return func(o *Options) {
		o.Response = response
	}
}

type Controller struct {
	taskDefines map[string]Task

	endpoints map[string]*types.Endpoint

	mu    sync.RWMutex
	pools map[string]Task
}

func NewController() *Controller {
	controller := &Controller{
		taskDefines: make(map[string]Task),
		pools:       make(map[string]Task),
	}
	return controller
}

func (c *Controller) ListEndpoints() []*types.Endpoint {
	endpoints := make([]*types.Endpoint, 0)
	for _, item := range c.endpoints {
		endpoints = append(endpoints, item.DeepCopy())
	}
	return endpoints
}

func (c *Controller) Register(task Task, opts ...Option) error { return nil }

func (c *Controller) RegisterFn(fn any, opts ...Option) error {
	endpoint, taskForFn, ok := extractFunc(fn, opts...)
	if !ok {
		return fmt.Errorf("invalid function")
	}

	name := endpoint.Name
	_, exists := c.endpoints[name]
	if exists {
		return fmt.Errorf("endpoint '%s' already exists", name)
	}
	c.endpoints[name] = endpoint

	c.taskDefines[name] = taskForFn
	return nil
}

func (r *Runner) handle(ctx context.Context, req *types.CallTaskRequest) (*types.CallTaskResponse, error) {
	lg := r.lg
	timeout := time.Duration(req.Timeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	results := make(map[string]*types.Value)
	results["a"] = types.NewValue("bb")

	resp := &types.CallTaskResponse{
		SeqId:   req.SeqId,
		Results: results,
	}

	lg.Info("call response", zap.Any("request", req))

	return resp, fmt.Errorf("internal error")
}

func doRecover() error {
	var err error
	if e := recover(); e != nil {
		err = fmt.Errorf("panic: %v", e)
	}
	return err
}
