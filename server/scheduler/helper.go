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

package scheduler

import (
	"github.com/olive-io/bpmn/schema"
	"github.com/olive-io/bpmn/v2"

	"github.com/olive-io/gflow/api/types"
)

func toSchemaValue(value *types.Value) *schema.Value {
	sv := &schema.Value{ItemValue: value.Value}
	switch value.Type {
	case types.Value_String:
		sv.ItemType = schema.ItemTypeString
	case types.Value_Integer:
		sv.ItemType = schema.ItemTypeInteger
	case types.Value_Boolean:
		sv.ItemType = schema.ItemTypeBoolean
	case types.Value_Float:
		sv.ItemType = schema.ItemTypeFloat
	case types.Value_Array:
		sv.ItemType = schema.ItemTypeArray
	case types.Value_Object:
		sv.ItemType = schema.ItemTypeObject
	}
	return sv
}

func fromSchemaValue(sv *schema.Value) *types.Value {
	tv := &types.Value{
		Value: sv.ItemValue,
	}
	switch sv.ItemType {
	case schema.ItemTypeString:
		tv.Type = types.Value_String
	case schema.ItemTypeInteger:
		tv.Type = types.Value_Integer
	case schema.ItemTypeBoolean:
		tv.Type = types.Value_Boolean
	case schema.ItemTypeFloat:
		tv.Type = types.Value_Float
	case schema.ItemTypeArray:
		tv.Type = types.Value_Array
	case schema.ItemTypeObject:
		tv.Type = types.Value_Object
	}
	return tv
}

func parseTaskType(at bpmn.ActivityType) types.FlowNodeType {
	switch at {
	case bpmn.TaskActivity:
		return types.FlowNodeType_Task
	case bpmn.BusinessRuleActivity:
		return types.FlowNodeType_BusinessRuleTask
	case bpmn.CallActivity:
		return types.FlowNodeType_CallActivity
	case bpmn.ManualTaskActivity:
		return types.FlowNodeType_ManualTask
	case bpmn.ScriptTaskActivity:
		return types.FlowNodeType_ScriptTask
	case bpmn.SendTaskActivity:
		return types.FlowNodeType_SendTask
	case bpmn.ServiceTaskActivity:
		return types.FlowNodeType_ServiceTask
	case bpmn.SubprocessActivity:
		return types.FlowNodeType_SubProcess
	case bpmn.ReceiveTaskActivity:
		return types.FlowNodeType_ReceiveTask
	case bpmn.UserTaskActivity:
		return types.FlowNodeType_UserTask
	default:
		return types.FlowNodeType_UserTask
	}
}

func parseElementType(elem schema.FlowElementInterface) types.FlowNodeType {
	switch elem.(type) {
	case *schema.StartEvent:
		return types.FlowNodeType_StartEvent
	case *schema.EndEvent:
		return types.FlowNodeType_EndEvent
	case *schema.CatchEvent:
		return types.FlowNodeType_IntermediateCatchEvent
	case *schema.BoundaryEvent:
		return types.FlowNodeType_BoundaryEvent
	case *schema.ParallelGateway:
		return types.FlowNodeType_ParallelGateway
	case *schema.InclusiveGateway:
		return types.FlowNodeType_InclusiveGateway
	case *schema.ExclusiveGateway:
		return types.FlowNodeType_ExclusiveGateway
	case *schema.EventBasedGateway:
		return types.FlowNodeType_EventBasedGateway
	case *schema.Task:
		return types.FlowNodeType_Task
	case *schema.BusinessRuleTask:
		return types.FlowNodeType_BusinessRuleTask
	case *schema.CallActivity:
		return types.FlowNodeType_CallActivity
	case *schema.ManualTask:
		return types.FlowNodeType_ManualTask
	case *schema.SendTask:
		return types.FlowNodeType_SendTask
	case *schema.ScriptTask:
		return types.FlowNodeType_ScriptTask
	case *schema.ServiceTask:
		return types.FlowNodeType_ServiceTask
	case *schema.SubProcess:
		return types.FlowNodeType_SubProcess
	case *schema.ReceiveTask:
		return types.FlowNodeType_ReceiveTask
	case *schema.UserTask:
		return types.FlowNodeType_UserTask
	default:
		return types.FlowNodeType_UnknownNode
	}
}
