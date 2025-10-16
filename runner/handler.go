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
	"time"

	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
)

var (
	originKey = struct{}{}
)

// GetOriginData returns Origin Parameter from gflow server
func GetOriginData(ctx context.Context) (*plugins.Request, bool) {
	value := ctx.Value(originKey)
	if value == nil {
		return nil, false
	}
	req, ok := value.(*plugins.Request)
	return req, ok
}

type TaskClone interface {
	Clone() TaskClone
}

type Task interface {
	Commit(ctx context.Context, request any) (any, error)
	Rollback(ctx context.Context) error
	Destroy(ctx context.Context) error
	String() string
}

var _ Task = (*taskProxy)(nil)

type taskProxy struct {
	opt   *Options
	proxy Task
}

func (tp *taskProxy) Clone() *taskProxy {
	out := new(taskProxy)
	out.proxy = tp.proxy
	out.opt = &Options{
		Name:     tp.opt.Name,
		Request:  tp.opt.Request,
		Response: tp.opt.Response,
	}
	return out
}

func (tp *taskProxy) Commit(ctx context.Context, in any) (any, error) {
	request := in.(*plugins.Request)

	argsType := reflect.TypeOf(tp.opt.Request)
	if argsType.Kind() == reflect.Ptr {
		argsType = argsType.Elem()
	}
	arg := reflect.New(argsType).Interface()
	if err := request.ApplyTo(arg); err != nil {
		return nil, err
	}

	call := func(ctx context.Context, req any) (resp any, err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()

		resp, err = tp.proxy.Commit(ctx, arg)
		return
	}

	out, err := call(ctx, request)
	if err != nil {
		return nil, err
	}

	resp := plugins.ExtractResponse(reflect.ValueOf(out))
	return resp, nil
}

func (tp *taskProxy) Rollback(ctx context.Context) error {
	call := func(ctx context.Context) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()

		err = tp.proxy.Rollback(ctx)
		return
	}

	return call(ctx)
}

func (tp *taskProxy) Destroy(ctx context.Context) error {
	call := func(ctx context.Context) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()

		err = tp.proxy.Destroy(ctx)
		return
	}

	return call(ctx)
}

func (tp *taskProxy) String() string { return tp.opt.Name }

var _ Task = (*fnProxy)(nil)

type fnProxy struct {
	name      string
	methodPtr reflect.Value
	args      []reflect.Type

	ctxIsFirst  bool
	containsReq bool
}

func (fn *fnProxy) Clone() Task {
	out := new(fnProxy)
	out.name = fn.name
	out.methodPtr = fn.methodPtr
	out.args = fn.args

	out.ctxIsFirst = fn.ctxIsFirst
	out.containsReq = fn.containsReq
	return out
}

func (fn *fnProxy) Commit(ctx context.Context, arg any) (any, error) {
	request := arg.(*plugins.Request)

	method := fn.methodPtr

	inputs := make([]reflect.Value, 0)
	if fn.ctxIsFirst {
		inputs = append(inputs, reflect.ValueOf(ctx))
	}
	if fn.containsReq {
		in := fn.args[0]
		target := reflect.New(in.Elem()).Interface()
		if err := request.ApplyTo(target); err != nil {
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

			inType := in
			if in.Kind() == reflect.Ptr {
				inType = in.Elem()
			}
			rv := reflect.New(inType).Elem()
			if err := tv.ApplyTo(rv); err != nil {
				return nil, err
			}
			inputs = append(inputs, rv)
		}
	}

	call := func(in []reflect.Value) (outs []reflect.Value, err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()
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

	resp := &plugins.Response{
		Results:     map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
	}
	switch len(outputs) {
	case 0:
	case 1:
		out := outputs[0]
		taskResp, ok := out.Interface().(*plugins.Response)
		if ok {
			return taskResp, nil
		}

		if out.Kind() == reflect.Ptr {
			out = out.Elem()
		}

		if out.Kind() == reflect.Struct {
			resp = plugins.ExtractResponse(out)
		} else {
			tv := types.FromReflectValue(out)
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

func (fn *fnProxy) Rollback(ctx context.Context) error { return nil }

func (fn *fnProxy) Destroy(ctx context.Context) error { return nil }

func (fn *fnProxy) String() string { return fn.name }

type Options struct {
	Name        string
	Type        types.FlowNodeType
	Kind        string
	Description string
	Request     any
	Response    any
}

type Option func(*Options)

func NewOptions(opts ...Option) *Options {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	if options.Type == 0 {
		options.Type = types.FlowNodeType_ServiceTask
	}
	if options.Kind == "" {
		options.Kind = "gflow"
	}

	return &options
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithType(t types.FlowNodeType) Option {
	return func(o *Options) {
		o.Type = t
	}
}

func WithKind(kind string) Option {
	return func(o *Options) {
		o.Kind = kind
	}
}

func WithDesc(desc string) Option {
	return func(o *Options) {
		o.Description = desc
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

func (r *Runner) ListEndpoints() []*types.Endpoint {
	endpoints := make([]*types.Endpoint, 0)
	for _, item := range r.endpoints {
		endpoints = append(endpoints, item.DeepCopy())
	}
	return endpoints
}

func (r *Runner) Register(task Task, opts ...Option) error {
	options := NewOptions(opts...)
	endpoint, proxy, ok := extractTask(task, options)
	if !ok {
		return fmt.Errorf("invalid task")
	}

	name := endpoint.Name
	_, exists := r.endpoints[name]
	if exists {
		return fmt.Errorf("endpoint '%s' already exists", name)
	}
	r.endpoints[name] = endpoint

	r.taskDefines[name] = proxy
	return nil
}

func (r *Runner) RegisterFn(fn any, opts ...Option) error {
	options := NewOptions(opts...)
	endpoint, proxy, ok := extractFunc(fn, options)
	if !ok {
		return fmt.Errorf("invalid function")
	}

	name := endpoint.Name
	_, exists := r.endpoints[name]
	if exists {
		return fmt.Errorf("endpoint '%s' already exists", name)
	}
	r.endpoints[name] = endpoint

	r.taskDefines[name] = proxy
	return nil
}

func (r *Runner) Handle(ctx context.Context, req *types.CallTaskRequest) *types.CallTaskResponse {
	lg := r.lg
	timeout := time.Duration(req.Timeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp := &types.CallTaskResponse{
		Stage:       req.Stage,
		SeqId:       req.SeqId,
		Results:     map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
	}

	ident := fmt.Sprintf("%d.%s", req.Process, req.Name)

	lg.Info("call task",
		zap.String("stage", req.Stage.String()),
		zap.String("task", req.Name))
	switch req.Stage {
	case types.CallTaskStage_Echo, types.CallTaskStage_Commit:
		stepCommitCounter.Add(1)
		defer stepCommitCounter.Sub(1)

		name := req.Name
		taskDef, found := r.taskDefines[name]
		if !found {
			resp.Error = fmt.Errorf("invalid kind '%s'", name).Error()
			return resp
		}

		task := taskDef
		if impl, ok := task.(TaskClone); ok {
			task = impl.Clone().(Task)
		}

		if req.Stage == types.CallTaskStage_Commit {
			r.pmu.Lock()
			r.pools[ident] = task
			r.pmu.Unlock()
		}

		in := &plugins.Request{
			Headers:     req.Headers,
			Properties:  req.Properties,
			DataObjects: req.DataObjects,
		}

		ctx = context.WithValue(ctx, originKey, in)
		out, err := task.Commit(ctx, in)
		if err != nil {
			resp.Error = err.Error()
			return resp
		}

		taskResp := out.(*plugins.Response)
		resp.Results = taskResp.Results
		resp.DataObjects = taskResp.DataObjects

	case types.CallTaskStage_Rollback:
		stepRollbackCounter.Add(1)
		defer stepRollbackCounter.Sub(1)

		r.pmu.RLock()
		task, ok := r.pools[ident]
		r.pmu.RUnlock()

		if ok {
			ctx = context.WithValue(ctx, originKey, req)
			if err := task.Rollback(ctx); err != nil {
				resp.Error = err.Error()
			}
		}

	case types.CallTaskStage_Destroy:
		stepDestroyCounter.Add(1)
		defer stepDestroyCounter.Sub(1)

		r.pmu.RLock()
		task, ok := r.pools[ident]
		r.pmu.RUnlock()

		if ok {
			ctx = context.WithValue(ctx, originKey, req)
			if err := task.Destroy(ctx); err != nil {
				resp.Error = err.Error()
			}

			r.pmu.Lock()
			delete(r.pools, ident)
			r.pmu.Unlock()
		}
	}

	return resp
}
