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
	"path"

	"github.com/olive-io/bpmn/schema"
)

func (in *Endpoint) URL() string {
	var urlText string
	if in.Kind != "" {
		urlText = path.Join(urlText, in.Kind)
	}
	if in.Name != "" {
		urlText = path.Join(urlText, in.Name)
	}
	return urlText
}

func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
}

func (in *Endpoint) DeepCopy() *Endpoint {
	if in == nil {
		return nil
	}
	out := new(Endpoint)
	in.DeepCopyInto(out)
	return out
}

func (in *Value) DeepCopyInto(out *Value) {
	*out = *in
}

func (in *Value) DeepCopy() *Value {
	if in == nil {
		return nil
	}
	out := new(Value)
	in.DeepCopyInto(out)
	return out
}

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

func (in *ProcessContext) DeepCopyInto(out *ProcessContext) {
	*out = *in
	for k, v := range in.DataObjects {
		out.DataObjects[k] = v.DeepCopy()
	}
	for k, v := range in.Variables {
		out.Variables[k] = v.DeepCopy()
	}
}

func (in *ProcessContext) DeepCopy() *ProcessContext {
	if in == nil {
		return nil
	}
	out := new(ProcessContext)
	in.DeepCopyInto(out)
	return out
}

func (in *Process) DeepCopyInto(out *Process) {
	*out = *in
	if in.Context != nil {
		in, out := in.Context, out.Context
		in.DeepCopyInto(out)
	}
}

func (in *Process) DeepCopy() *Process {
	if in == nil {
		return nil
	}
	out := new(Process)
	in.DeepCopyInto(out)
	return out
}

func (in *FlowNode) DeepCopyInto(out *FlowNode) {
	*out = *in
	if in.Headers != nil {
		for k, v := range in.Headers {
			out.Headers[k] = v
		}
	}
	if in.Properties != nil {
		for k, v := range in.Properties {
			out.Properties[k] = v.DeepCopy()
		}
	}
	if in.DataObjects != nil {
		for k, v := range in.DataObjects {
			out.DataObjects[k] = v.DeepCopy()
		}
	}
}

func (in *FlowNode) DeepCopy() *FlowNode {
	if in == nil {
		return nil
	}
	out := new(FlowNode)
	in.DeepCopyInto(out)
	return out
}

func (in *Runner) DeepCopyInto(out *Runner) {
	*out = *in
	if in.Metadata != nil {
		for k, v := range in.Metadata {
			out.Metadata[k] = v
		}
	}
	if in.Features != nil {
		for k, v := range in.Features {
			out.Features[k] = v
		}
	}
}

func (in *Runner) DeepCopy() *Runner {
	if in == nil {
		return nil
	}
	out := new(Runner)
	in.DeepCopyInto(out)
	return out
}

func (in *RunnerStat) DeepCopyInto(out *RunnerStat) {
	*out = *in
}

func (in *RunnerStat) DeepCopy() *RunnerStat {
	if in == nil {
		return nil
	}
	out := new(RunnerStat)
	in.DeepCopyInto(out)
	return out
}

// ConvertStage converts FlowNode_FlowNodeStage to CallTaskRequest_Stage
func ConvertStage(in FlowNode_FlowNodeStage) CallTaskStage {
	switch in {
	case FlowNode_Commit:
		return CallTaskStage_Commit
	case FlowNode_Rollback:
		return CallTaskStage_Rollback
	case FlowNode_Destroy:
		return CallTaskStage_Destroy
	default:
		return CallTaskStage_Commit
	}
}
