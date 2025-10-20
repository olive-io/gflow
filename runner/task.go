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
	"path/filepath"
	"reflect"
	"runtime"

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
	opt   *plugins.RegisterOptions
	proxy Task
}

func (tp *taskProxy) Clone() *taskProxy {
	out := new(taskProxy)
	out.proxy = tp.proxy
	out.opt = &plugins.RegisterOptions{
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

func extractTask(task Task, options *plugins.RegisterOptions) (*types.Endpoint, *taskProxy, error) {
	if options.Request == nil {
		return nil, nil, fmt.Errorf("no request provided")
	}
	if options.Response == nil {
		return nil, nil, fmt.Errorf("no response provided")
	}

	endpoint := &types.Endpoint{
		TaskType:    options.FlowType,
		Type:        options.Type,
		Description: options.Description,
		Mode:        types.TransitionMode_Transition,
		Metadata:    map[string]string{},
		Headers:     map[string]string{},
		Properties:  map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
		Results:     map[string]*types.Value{},
	}

	impl := &taskProxy{
		opt:   options,
		proxy: task,
	}

	rt := reflect.TypeOf(task)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	name := rt.PkgPath() + "." + rt.Name()
	endpoint.Metadata["fullName"] = name
	if options.Name != "" {
		name = options.Name
	}
	endpoint.Name = name

	headers, properties, dataObjects, matched := plugins.ExtractInOrOut(options.Request)
	if !matched {
		return nil, nil, fmt.Errorf("bad request")
	}
	endpoint.Headers = headers
	endpoint.Properties = properties
	endpoint.DataObjects = dataObjects

	_, results, _, matched := plugins.ExtractInOrOut(options.Response)
	if !matched {
		return nil, nil, fmt.Errorf("bad response")
	}
	endpoint.Results = results

	return endpoint, impl, nil
}

func extractFunc(fn any, options *plugins.RegisterOptions) (*types.Endpoint, *fnProxy, error) {
	rv := reflect.ValueOf(fn)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		return nil, nil, fmt.Errorf("must be a function")
	}

	endpoint := &types.Endpoint{
		TaskType:    options.FlowType,
		Type:        options.Type,
		Description: options.Description,
		Metadata:    map[string]string{},
		Headers:     map[string]string{},
		Properties:  map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
		Results:     map[string]*types.Value{},
	}

	taskFn := &fnProxy{
		methodPtr:   rv,
		args:        make([]reflect.Type, 0),
		ctxIsFirst:  false,
		containsReq: false,
	}

	pc := runtime.FuncForPC(rv.Pointer())
	endpoint.Metadata["fullName"] = pc.Name()
	name := filepath.Base(pc.Name())
	if options.Name != "" {
		name = options.Name
	}
	endpoint.Name = name
	taskFn.name = name

	if options.Request != nil {
		headers, properties, dataObjects, matched := plugins.ExtractInOrOut(options.Request)
		if !matched {
			return nil, nil, fmt.Errorf("bad request")
		}
		endpoint.Headers = headers
		endpoint.Properties = properties
		endpoint.DataObjects = dataObjects

		taskFn.args = append(taskFn.args, reflect.TypeOf(options.Request))
	} else {
		switch rt.NumIn() {
		case 0:
		case 1:
			in := rt.In(0)
			if !isContext(in) {
				if isStruct(in) {
					inImpl := reflect.New(in.Elem()).Interface()
					headers, properties, dataObjects, matched := plugins.ExtractInOrOut(inImpl)
					if matched {
						taskFn.containsReq = true
						endpoint.Headers = headers
						endpoint.Properties = properties
						endpoint.DataObjects = dataObjects
					}
				} else {
					tv, ok := plugins.ExtractReflectType(in)
					if ok {
						endpoint.Properties["p0"] = tv
					}
				}
				taskFn.args = append(taskFn.args, in)
			} else {
				taskFn.ctxIsFirst = true
			}
		case 2:
			if isContext(rt.In(0)) {
				taskFn.ctxIsFirst = true
				in := rt.In(1)
				taskFn.args = append(taskFn.args, in)
				if isStruct(in) {
					inImpl := reflect.New(in.Elem()).Interface()
					headers, properties, dataObjects, matched := plugins.ExtractInOrOut(inImpl)
					if matched {
						taskFn.containsReq = true
						endpoint.Headers = headers
						endpoint.Properties = properties
						endpoint.DataObjects = dataObjects
					}
				} else {
					tv, ok := plugins.ExtractReflectType(in)
					if ok {
						endpoint.Properties["p0"] = tv
					}
				}
			} else {
				p := 0
				for i := 0; i < rt.NumIn(); i++ {
					in := rt.In(i)
					taskFn.args = append(taskFn.args, in)

					tv, ok := plugins.ExtractReflectType(in)
					if ok {
						endpoint.Properties[fmt.Sprintf("p%d", p)] = tv
						p += 1
					}
				}
			}
		default:
			p := 0
			for i := 0; i < rt.NumIn(); i++ {
				in := rt.In(i)
				if isContext(in) {
					continue
				}

				taskFn.args = append(taskFn.args, in)
				tv, ok := plugins.ExtractReflectType(in)
				if ok {
					endpoint.Properties[fmt.Sprintf("p%d", p)] = tv
					p += 1
				}
			}
		}
	}

	if options.Response != nil {
		_, properties, _, matched := plugins.ExtractInOrOut(options.Response)
		if !matched {
			return nil, nil, fmt.Errorf("bad response")
		}
		endpoint.Results = properties
	} else {
		switch rt.NumOut() {
		case 0:
		case 1:
			if !isErr(rt.Out(0)) {
				out := rt.Out(0)
				if isStruct(out) {
					outImpl := reflect.New(out).Interface()
					_, properties, _, matched := plugins.ExtractInOrOut(outImpl)
					if matched {
						endpoint.Results = properties
					}
				} else {
					tv, ok := plugins.ExtractReflectType(out)
					if ok {
						endpoint.Results["r0"] = tv
					}
				}
			}
		case 2:
			if isErr(rt.Out(1)) {
				out := rt.Out(0)
				if isStruct(out) {
					outImpl := reflect.New(out).Interface()
					_, properties, _, matched := plugins.ExtractInOrOut(outImpl)
					if matched {
						endpoint.Results = properties
					}
				} else {
					tv, ok := plugins.ExtractReflectType(out)
					if ok {
						endpoint.Results["r0"] = tv
					}
				}
			} else {
				p := 0
				for i := 0; i < rt.NumOut(); i++ {
					out := rt.Out(i)
					if isErr(out) {
						continue
					}

					tv, ok := plugins.ExtractReflectType(out)
					if ok {
						key := fmt.Sprintf("r%d", p)
						endpoint.Results[key] = tv
						p += 1
					}
				}
			}
		default:
			p := 0
			for i := 0; i < rt.NumOut(); i++ {
				out := rt.Out(i)
				if isErr(out) {
					continue
				}

				tv, ok := plugins.ExtractReflectType(out)
				if ok {
					key := fmt.Sprintf("r%d", p)
					endpoint.Results[key] = tv
					p += 1
				}
			}
		}
	}

	return endpoint, taskFn, nil
}

func isContext(rt reflect.Type) bool {
	return rt == reflect.TypeOf((*context.Context)(nil)).Elem()
}

func isErr(rt reflect.Type) bool {
	return rt == reflect.TypeOf((*error)(nil)).Elem()
}

func isStruct(rt reflect.Type) bool {
	if rt.Kind() == reflect.Ptr {
		return isStruct(rt.Elem())
	}
	return rt.Kind() == reflect.Struct
}
