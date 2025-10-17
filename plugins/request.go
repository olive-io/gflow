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

package plugins

import (
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

func (req *Request) ApplyTo(target any) error {
	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}
	if rv, ok := target.(reflect.Value); ok {
		return req.ApplyToReflectValue(rv)
	}
	return req.ApplyToReflectValue(reflect.ValueOf(target))
}

func (req *Request) ApplyToReflectValue(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Ptr:
		return req.ApplyToReflectValue(rv.Elem())
	case reflect.Struct:
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			ft := rt.Field(i)
			fv := rv.Field(i)
			if !ft.IsExported() {
				continue
			}

			tag, found := ft.Tag.Lookup("json")
			if found {
				key := strings.Split(tag, ",")[0]
				tv, ok := req.Properties[key]
				if ok {
					if err := tv.ApplyTo(fv); err != nil {
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
						if err := tv.ApplyTo(fv); err != nil {
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
		return tv.ApplyTo(rv)
	}

	return nil
}

func ExtractResponse(rv reflect.Value) *Response {
	resp := &Response{
		Results:     make(map[string]*types.Value),
		DataObjects: make(map[string]*types.Value),
	}
	rt := rv.Type()
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return resp
	}
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		fv := rv.Field(i)

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
				resp.DataObjects[key] = tv
			}
		}
	}

	return resp
}

func ExtractInOrOut(in any) (headers map[string]string, properties map[string]*types.Value, dataObjects map[string]*types.Value, ok bool) {
	headers = make(map[string]string)
	properties = make(map[string]*types.Value)
	dataObjects = make(map[string]*types.Value)

	rv := reflect.ValueOf(in)
	rt := rv.Type()
	if rt.Kind() == reflect.Ptr {
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
	if rt.Kind() == reflect.Ptr {
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
