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

package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/olive-io/bpmn/schema"
)

func NewValue(value any) *Value {
	return FromSchemaValue(schema.NewValue(value))
}

func FromSchemaValue(sv *schema.Value) *Value {
	value := &Value{
		Type:  fromSchemaType(sv.ItemType),
		Value: sv.ItemValue,
	}
	return value
}

func (in *Value) ToSchemaValue() *schema.Value {
	sv := &schema.Value{
		ItemType:  toSchemaType(in.Type),
		ItemValue: in.Value,
	}
	return sv
}

func (in *Value) ValueFrom(value any) {
	vv, ok := value.(*Value)
	if ok {
		*in = *vv
		return
	}

	out := FromReflectValue(reflect.ValueOf(value))
	*in = *out
}

func (in *Value) ValueTo(target any) error {
	return in.ToSchemaValue().ValueTo(target)
}

func (in *Value) ApplyTo(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.String:
		if in.Type != Value_String {
			return fmt.Errorf("value types must be %s", in.Type.String())
		}
		rv.SetString(in.Value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if in.Type != Value_Integer {
			return fmt.Errorf("value types must be %s", in.Type.String())
		}
		n, err := strconv.ParseInt(in.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("convert value %s to int: %v", in.Value, err)
		}
		rv.SetInt(n)
	case reflect.Float32, reflect.Float64:
		if in.Type != Value_Float {
			return fmt.Errorf("value types must be %s", in.Type.String())
		}
		n, err := strconv.ParseFloat(in.Value, 64)
		if err != nil {
			return fmt.Errorf("convert value %s to float: %v", in.Value, err)
		}
		rv.SetFloat(n)
	case reflect.Bool:
		if in.Type != Value_Boolean {
			return fmt.Errorf("value types must be %s", in.Type.String())
		}
		n, err := strconv.ParseBool(in.Value)
		if err != nil {
			return fmt.Errorf("convert value %s to bool: %v", in.Value, err)
		}
		rv.SetBool(n)
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 && in.Type == Value_String {
			rv.SetBytes([]byte(in.Value))
		} else {
			if in.Type != Value_Array {
				return fmt.Errorf("value types must be %s", in.Type.String())
			}
			v := reflect.New(rv.Type())
			vv := v.Interface()

			err := json.Unmarshal([]byte(in.Value), vv)
			if err != nil {
				return err
			}
			rv.Set(v.Elem())
		}
	case reflect.Map, reflect.Struct:
		if in.Type != Value_Object {
			return fmt.Errorf("value types must be %s", in.Type.String())
		}
		v := reflect.New(rv.Type())
		vv := v.Interface()

		err := json.Unmarshal([]byte(in.Value), &vv)
		if err != nil {
			return err
		}
		rv.Set(reflect.Indirect(v))
	case reflect.Ptr:
		v := reflect.New(rv.Type().Elem())
		vv := v.Interface()

		err := json.Unmarshal([]byte(in.Value), &vv)
		if err != nil {
			return err
		}
		rv.Set(v)
	default:
		return fmt.Errorf("invalid type %s", rv.Type().Name())
	}

	return nil
}

func FromReflectValue(rv reflect.Value) *Value {
	switch rv.Kind() {
	case reflect.String:
		return NewValue(rv.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewValue(rv.Int())
	case reflect.Float32, reflect.Float64:
		return NewValue(rv.Float())
	case reflect.Bool:
		return NewValue(rv.Bool())
	case reflect.Slice:
		v := reflect.New(rv.Type().Elem())
		vv := v.Interface()
		v.Set(rv)
		return NewValue(vv)
	case reflect.Struct, reflect.Map:
		v := reflect.New(rv.Type().Elem())
		vv := v.Interface()
		v.Set(rv)
		return NewValue(vv)
	case reflect.Ptr:
		return FromReflectValue(rv.Elem())
	default:
		return NewValue("")
	}
}

func fromSchemaType(st schema.ItemType) Value_Type {
	switch st {
	case schema.ItemTypeString:
		return Value_String
	case schema.ItemTypeInteger:
		return Value_Integer
	case schema.ItemTypeFloat:
		return Value_Float
	case schema.ItemTypeBoolean:
		return Value_Boolean
	case schema.ItemTypeArray:
		return Value_Array
	case schema.ItemTypeObject:
		return Value_Object
	default:
		return Value_String
	}
}

func toSchemaType(vt Value_Type) schema.ItemType {
	switch vt {
	case Value_String:
		return schema.ItemTypeString
	case Value_Integer:
		return schema.ItemTypeInteger
	case Value_Float:
		return schema.ItemTypeFloat
	case Value_Boolean:
		return schema.ItemTypeBoolean
	case Value_Array:
		return schema.ItemTypeArray
	case Value_Object:
		return schema.ItemTypeObject
	default:
		return schema.ItemTypeString
	}
}
