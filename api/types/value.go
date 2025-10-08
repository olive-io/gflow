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
	"github.com/olive-io/bpmn/schema"
)

func ToSchemaType(in Value_Type) schema.ItemType {
	switch in {
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

func FromSchemaType(in schema.ItemType) Value_Type {
	switch in {
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

func (in *Value) ToSchemaValue() *schema.Value {
	sv := &schema.Value{
		ItemType:  ToSchemaType(in.Type),
		ItemValue: in.Value,
	}
	return sv
}

func FromSchemaValue(sv *schema.Value) *Value {
	v := &Value{
		Type:  FromSchemaType(sv.ItemType),
		Value: sv.ItemValue,
	}
	return v
}

func NewValue(value any) *Value {
	return FromSchemaValue(schema.NewValue(value))
}
