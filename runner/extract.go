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

	"github.com/olive-io/gflow/api/types"
)

var (
	DefaultTag    = "gflow"
	HeaderTag     = "hr" // task header tag
	DataObjectTag = "dt" // task data object tag
)

func ExtractTask(task Task, opts ...Option) (*types.Endpoint, bool) {
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
		Name:        task.String(),
		Headers:     map[string]string{},
		Properties:  map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
		Results:     map[string]*types.Value{},
	}

	headers, properties, dataObjects, matched := ExtractInOrOut(options.Request)
	if !matched {
		return nil, false
	}
	endpoint.Headers = headers
	endpoint.Properties = properties
	endpoint.DataObjects = dataObjects

	_, results, _, matched := ExtractInOrOut(options.Request)
	if !matched {
		return nil, false
	}
	endpoint.Results = results

	return endpoint, true
}

func ExtractFunc(fn any, opts ...Option) (*types.Endpoint, bool) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	rt := reflect.TypeOf(fn)
	if rt.Kind() != reflect.Func {
		return nil, false
	}

	name := rt.PkgPath() + "." + rt.Name()
	if options.Name != "" {
		name = options.Name
	}

	endpoint := &types.Endpoint{
		Name:        name,
		Headers:     map[string]string{},
		Properties:  map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
		Results:     map[string]*types.Value{},
	}

	if options.Request != nil {
		headers, properties, dataObjects, matched := ExtractInOrOut(options.Request)
		if !matched {
			return nil, false
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
				tv, ok := ExtractReflectType(in)
				if ok {
					endpoint.Properties["p1"] = tv
				}
			}
		case 2:
			if isContext(rt.In(0)) {
				defaultKey := "in"
				values := ExtractDepthValue(rt.In(1), defaultKey)
				for k, v := range values {
					endpoint.Properties[k] = v
				}
			} else {
				p := 0
				for i := 0; i < rt.NumIn(); i++ {
					in := rt.In(i)

					tv, ok := ExtractReflectType(in)
					if ok {
						p += 1
						endpoint.Properties[fmt.Sprintf("p%d", p)] = tv
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
					p += 1
					endpoint.Properties[fmt.Sprintf("p%d", p)] = tv
				}
			}
		}
	}

	if options.Response != nil {
		_, properties, _, matched := ExtractInOrOut(options.Response)
		if !matched {
			return nil, false
		}
		endpoint.Results = properties
	} else {
		switch rt.NumOut() {
		case 0:
		case 1:
			if !isErr(rt.Out(0)) {
				key := "out"
				values := ExtractDepthValue(rt.Out(0), key)
				for k, v := range values {
					endpoint.Results[k] = v
				}
			}
		case 2:
			if isErr(rt.Out(1)) {
				key := "out"
				values := ExtractDepthValue(rt.Out(0), key)
				for k, v := range values {
					endpoint.Results[k] = v
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
					p += 1
					endpoint.Properties[fmt.Sprintf("out%d", p)] = tv
				}
			}
		}
	}

	return endpoint, true
}

func ExtractDepthValue(rt reflect.Type, defaultKey string) map[string]*types.Value {
	name := defaultKey
	values := map[string]*types.Value{}
	switch rt.Kind() {
	case reflect.String:
		values[name] = &types.Value{Type: types.Value_String}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		values[name] = &types.Value{Type: types.Value_Integer}
	case reflect.Float32, reflect.Float64:
		values[name] = &types.Value{Type: types.Value_Float}
	case reflect.Bool:
		values[name] = &types.Value{Type: types.Value_Boolean}
	case reflect.Array, reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 {
			values[name] = &types.Value{Type: types.Value_String}
		} else {
			values[name] = &types.Value{Type: types.Value_Array}
		}
	case reflect.Struct:
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			fTag := field.Tag.Get("json")

			key := strings.Split(fTag, ",")[0]
			tv, ok := ExtractReflectType(field.Type)
			if ok {
				values[key] = tv
			}
		}
	case reflect.Map:
		values[name] = &types.Value{Type: types.Value_Object}
	default:
	}

	return values
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
	case reflect.Struct, reflect.Map:
		tv := &types.Value{Type: types.Value_Object}
		return tv, true
	default:
		return nil, false
	}
}

func isContext(rt reflect.Type) bool {
	return rt == reflect.TypeOf((*context.Context)(nil)).Elem()
}

func isErr(rt reflect.Type) bool {
	return !rt.Implements(reflect.TypeOf((*error)(nil)).Elem())
}
