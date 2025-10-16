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

func extractTask(task Task, options *Options) (*types.Endpoint, *taskProxy, bool) {
	if options.Request == nil {
		return nil, nil, false
	}
	if options.Response == nil {
		return nil, nil, false
	}

	endpoint := &types.Endpoint{
		Kind:          options.Kind,
		Type:          options.Type,
		Description:   options.Description,
		OnTransaction: true,
		Metadata:      map[string]string{},
		Headers:       map[string]string{},
		Properties:    map[string]*types.Value{},
		DataObjects:   map[string]*types.Value{},
		Results:       map[string]*types.Value{},
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
		return nil, nil, false
	}
	endpoint.Headers = headers
	endpoint.Properties = properties
	endpoint.DataObjects = dataObjects

	_, results, _, matched := plugins.ExtractInOrOut(options.Response)
	if !matched {
		return nil, nil, false
	}
	endpoint.Results = results

	return endpoint, impl, true
}

func extractFunc(fn any, options *Options) (*types.Endpoint, *fnProxy, bool) {
	rv := reflect.ValueOf(fn)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		return nil, nil, false
	}

	endpoint := &types.Endpoint{
		Kind:        options.Kind,
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
			return nil, nil, false
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
			return nil, nil, false
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

	return endpoint, taskFn, true
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
