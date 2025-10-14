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
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/olive-io/gflow/api/types"
)

var (
	DefaultTag    = "gflow"
	HeaderTag     = "hr" // task header tag
	DataObjectTag = "dt" // task data object tag
)

type TaskRequest struct {
	Headers     map[string]string
	Properties  map[string]*types.Value
	DataObjects map[string]*types.Value
}

func (req *TaskRequest) InjectFor(target any) error {
	return req.InjectForReflectValue(reflect.ValueOf(target))
}

func (req *TaskRequest) InjectForReflectValue(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Pointer:
		return req.InjectFor(rv.Elem())
	case reflect.Struct:
		rt := rv.Type().Elem()
		for i := 0; i < rt.NumField(); i++ {
			ft := rt.Field(i)
			fv := rv.Elem().Field(i)
			if !ft.IsExported() {
				continue
			}

			tag, found := ft.Tag.Lookup("json")
			if found {
				key := strings.Split(tag, ",")[0]
				value, ok := req.Properties[key]
				if ok {
					if err := InjectFromTypesValue(fv, value); err != nil {
						return fmt.Errorf("inject field '%s': %w", ft.Name, err)
					}
				}
				continue
			}

			tag, found = ft.Tag.Lookup("gflow")
			if found {
				parts := strings.Split(tag, ";")
				pairs := map[string]string{}
				for _, part := range parts {
					kv := strings.Split(part, ":")
					if len(kv) != 2 {
						continue
					}
					pairs[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}

				if key, exists := pairs[HeaderTag]; exists {
					if fv.Kind() == reflect.String {
						value, ok := req.Headers[key]
						if ok {
							fv.SetString(value)
						}
					}
				}

				if key, exists := pairs[DataObjectTag]; exists {
					tv, ok := req.DataObjects[key]
					if ok {
						if err := InjectFromTypesValue(fv, tv); err != nil {
							return fmt.Errorf("inject field '%s': %w", ft.Name, err)
						}
					}
				}
			}
		}
	default:
		tv, ok := req.Properties["p0"]
		if !ok {
			return fmt.Errorf("missing 'p0' property")
		}
		return InjectFromTypesValue(rv, tv)
	}

	return nil
}

type TaskResponse struct {
	Results     map[string]*types.Value
	DataObjects map[string]*types.Value
}

func extractTask(task Task, opts ...Option) (*types.Endpoint, bool) {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	if options.Request == nil {
		return nil, false
	}
	if options.Response == nil {
		return nil, false
	}

	endpoint := &types.Endpoint{
		Metadata:    map[string]string{},
		Headers:     map[string]string{},
		Properties:  map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
		Results:     map[string]*types.Value{},
	}

	rt := reflect.TypeOf(task)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}

	name := rt.PkgPath() + "." + rt.Name()
	endpoint.Metadata["pkgName"] = name
	if options.Name != "" {
		name = options.Name
	}
	endpoint.Name = name

	headers, properties, dataObjects, matched := ExtractInOrOut(options.Request)
	if !matched {
		return nil, false
	}
	endpoint.Headers = headers
	endpoint.Properties = properties
	endpoint.DataObjects = dataObjects

	_, results, _, matched := ExtractInOrOut(options.Response)
	if !matched {
		return nil, false
	}
	endpoint.Results = results

	return endpoint, true
}

func extractFunc(fn any, opts ...Option) (*types.Endpoint, *taskForFunc, bool) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	rv := reflect.ValueOf(fn)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		return nil, nil, false
	}

	endpoint := &types.Endpoint{
		Metadata:    map[string]string{},
		Headers:     map[string]string{},
		Properties:  map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
		Results:     map[string]*types.Value{},
	}

	taskFn := &taskForFunc{
		methodPtr:   rv,
		args:        make([]reflect.Type, 0),
		ctxIsFirst:  false,
		containsReq: false,
	}

	pc := runtime.FuncForPC(rv.Pointer())
	endpoint.Metadata["pkgName"] = pc.Name()
	name := filepath.Base(pc.Name())
	if options.Name != "" {
		name = options.Name
	}
	endpoint.Name = name
	taskFn.name = name

	if options.Request != nil {
		headers, properties, dataObjects, matched := ExtractInOrOut(options.Request)
		if !matched {
			return nil, nil, false
		}
		endpoint.Headers = headers
		endpoint.Properties = properties
		endpoint.DataObjects = dataObjects
	} else {
		switch rt.NumIn() {
		case 0:
		case 1:
			in := rt.In(0)
			if !isContext(in) {
				if isStruct(in) {
					inImpl := reflect.New(in.Elem()).Interface()
					headers, properties, dataObjects, matched := ExtractInOrOut(inImpl)
					if matched {
						taskFn.containsReq = true
						endpoint.Headers = headers
						endpoint.Properties = properties
						endpoint.DataObjects = dataObjects
					}
				} else {
					tv, ok := ExtractReflectType(in)
					if ok {
						endpoint.Properties["p0"] = tv
					}
				}
			} else {
				taskFn.ctxIsFirst = true
			}
		case 2:
			if isContext(rt.In(0)) {
				taskFn.ctxIsFirst = true
				in := rt.In(1)
				if isStruct(in) {
					inImpl := reflect.New(in.Elem()).Interface()
					headers, properties, dataObjects, matched := ExtractInOrOut(inImpl)
					if matched {
						taskFn.containsReq = true
						endpoint.Headers = headers
						endpoint.Properties = properties
						endpoint.DataObjects = dataObjects
					}
				} else {
					tv, ok := ExtractReflectType(in)
					if ok {
						endpoint.Properties["p0"] = tv
					}
				}
			} else {
				p := 0
				for i := 0; i < rt.NumIn(); i++ {
					in := rt.In(i)

					tv, ok := ExtractReflectType(in)
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

				tv, ok := ExtractReflectType(in)
				if ok {
					endpoint.Properties[fmt.Sprintf("p%d", p)] = tv
					p += 1
				}
			}
		}
	}

	if options.Response != nil {
		_, properties, _, matched := ExtractInOrOut(options.Response)
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
					_, properties, _, matched := ExtractInOrOut(outImpl)
					if matched {
						endpoint.Results = properties
					}
				} else {
					tv, ok := ExtractReflectType(out)
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
					_, properties, _, matched := ExtractInOrOut(outImpl)
					if matched {
						endpoint.Results = properties
					}
				} else {
					tv, ok := ExtractReflectType(out)
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

					tv, ok := ExtractReflectType(out)
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

				tv, ok := ExtractReflectType(out)
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

func ExtractInOrOut(in any) (map[string]string, map[string]*types.Value, map[string]*types.Value, bool) {
	headers := make(map[string]string)
	properties := make(map[string]*types.Value)
	dataObjects := make(map[string]*types.Value)

	rv := reflect.ValueOf(in)
	rt := rv.Type()
	if rt.Kind() == reflect.Pointer {
		rv = rv.Elem()
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return nil, nil, nil, false
	}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		fTag := field.Tag.Get(DefaultTag)
		if fTag == "" {
			fTag = field.Tag.Get("json")
			name := strings.Split(fTag, ",")[0]
			tv, ok := ExtractReflectType(field.Type)
			if ok {
				properties[name] = tv
			}
			continue
		}

		parts := strings.Split(fTag, ";")
		pairs := map[string]string{}
		for _, part := range parts {
			pair := strings.Split(part, ":")
			if len(pair) != 2 {
				continue
			}
			key := strings.TrimSpace(pair[0])
			value := strings.TrimSpace(pair[1])
			pairs[key] = value
		}

		tv, matched := ExtractReflectType(field.Type)
		if !matched {
			continue
		}

		if v, ok := pairs[HeaderTag]; ok && field.Type.Kind() == reflect.String {
			headers[v] = ""
		}
		if v, ok := pairs[DataObjectTag]; ok {
			dataObjects[v] = tv
		}
	}

	return headers, properties, dataObjects, true
}

func ExtractReflectType(rt reflect.Type) (*types.Value, bool) {
	if rt.Kind() == reflect.Pointer {
		return ExtractReflectType(rt.Elem())
	}

	switch rt.Kind() {
	case reflect.String:
		tv := &types.Value{Type: types.Value_String}
		return tv, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		tv := &types.Value{Type: types.Value_Integer}
		return tv, true
	case reflect.Float32, reflect.Float64:
		tv := &types.Value{Type: types.Value_Float}
		return tv, true
	case reflect.Bool:
		tv := &types.Value{Type: types.Value_Boolean}
		return tv, true
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			tv := types.NewValue("")
			return tv, true
		}
		tv := &types.Value{Type: types.Value_Array}
		return tv, true
	case reflect.Struct:
		tv := &types.Value{Type: types.Value_Object}
		tv.Kind = rt.PkgPath() + "." + rt.Name()
		return tv, true
	case reflect.Map:
		tv := &types.Value{Type: types.Value_Object, Kind: "map"}
		return tv, true
	default:
		return nil, false
	}
}

func InjectFromTypesValue(rv reflect.Value, value *types.Value) error {
	switch rv.Kind() {
	case reflect.String:
		if value.Type != types.Value_String {
			return fmt.Errorf("value types must be %s", value.Type.String())
		}
		rv.SetString(value.Value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value.Type != types.Value_Integer {
			return fmt.Errorf("value types must be %s", value.Type.String())
		}
		n, err := strconv.ParseInt(value.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("convert value %s to int: %v", value.Value, err)
		}
		rv.SetInt(n)
	case reflect.Float32, reflect.Float64:
		if value.Type != types.Value_Float {
			return fmt.Errorf("value types must be %s", value.Type.String())
		}
		n, err := strconv.ParseFloat(value.Value, 64)
		if err != nil {
			return fmt.Errorf("convert value %s to float: %v", value.Value, err)
		}
		rv.SetFloat(n)
	case reflect.Bool:
		if value.Type != types.Value_Boolean {
			return fmt.Errorf("value types must be %s", value.Type.String())
		}
		n, err := strconv.ParseBool(value.Value)
		if err != nil {
			return fmt.Errorf("convert value %s to bool: %v", value.Value, err)
		}
		rv.SetBool(n)
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 && value.Type == types.Value_String {
			rv.SetBytes([]byte(value.Value))
		} else {
			if value.Type != types.Value_Array {
				return fmt.Errorf("value types must be %s", value.Type.String())
			}
			v := reflect.New(rv.Type())
			vv := v.Interface()

			err := json.Unmarshal([]byte(value.Value), vv)
			if err != nil {
				return err
			}
			rv.Set(v.Elem())
		}
	case reflect.Map, reflect.Struct:
		if value.Type != types.Value_Object {
			return fmt.Errorf("value types must be %s", value.Type.String())
		}
		v := reflect.New(rv.Type())
		vv := v.Interface()

		err := json.Unmarshal([]byte(value.Value), &vv)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(&vv))
	case reflect.Pointer:
		v := reflect.New(rv.Type().Elem())
		vv := v.Interface()

		err := json.Unmarshal([]byte(value.Value), &vv)
		if err != nil {
			return err
		}
		rv.Set(v)
	default:
		return fmt.Errorf("invalid type %s", rv.Type().Name())
	}

	return nil
}

func fromReflectValue(rv reflect.Value) *types.Value {
	switch rv.Kind() {
	case reflect.String:
		return types.NewValue(rv.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return types.NewValue(rv.Int())
	case reflect.Float32, reflect.Float64:
		return types.NewValue(rv.Float())
	case reflect.Bool:
		return types.NewValue(rv.Bool())
	case reflect.Slice:
		v := reflect.New(rv.Type().Elem())
		vv := v.Interface()
		v.Set(rv)
		return types.NewValue(vv)
	case reflect.Struct, reflect.Map:
		v := reflect.New(rv.Type().Elem())
		vv := v.Interface()
		v.Set(rv)
		return types.NewValue(vv)
	case reflect.Ptr:
		return fromReflectValue(rv.Elem())
	default:
		return types.NewValue("")
	}
}

func isContext(rt reflect.Type) bool {
	return rt == reflect.TypeOf((*context.Context)(nil)).Elem()
}

func isErr(rt reflect.Type) bool {
	return rt == reflect.TypeOf((*error)(nil)).Elem()
}

func isStruct(rt reflect.Type) bool {
	if rt.Kind() == reflect.Pointer {
		return isStruct(rt.Elem())
	}
	return rt.Kind() == reflect.Struct
}
