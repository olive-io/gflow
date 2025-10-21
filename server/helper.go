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

package server

import (
	json "github.com/bytedance/sonic"
	"github.com/olive-io/bpmn/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/dao"
)

func toGRPCErr(err error) error {
	if dao.IsNotFound(err) {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func convertToTask(ep *types.Endpoint) schema.TaskInterface {
	headers := ep.Headers
	if headers == nil {
		headers = make(map[string]string)
	}
	headers[plugins.PrefixTag+"name"] = ep.Name

	extension := &schema.ExtensionElements{}
	metadata := ep.Metadata
	metadataBytes, _ := json.Marshal(metadata)
	extension.TaskDefinitionField = &schema.TaskDefinition{
		Type:     ep.Type,
		Metadata: string(metadataBytes),
	}

	taskHeader := &schema.TaskHeader{
		Header: make([]*schema.Item, 0),
	}
	for name, value := range headers {
		item := &schema.Item{
			Name:  name,
			Value: value,
			Type:  schema.ItemTypeString,
		}
		taskHeader.Header = append(taskHeader.Header, item)
	}
	extension.TaskHeaderField = taskHeader

	taskProperties := &schema.Properties{
		Property: make([]*schema.Item, 0),
	}
	for name, value := range ep.Properties {
		item := &schema.Item{
			Name:  name,
			Value: value.Value,
			Type:  types.ToSchemaType(value.Type),
		}
		taskProperties.Property = append(taskProperties.Property, item)
	}
	extension.PropertiesField = taskProperties

	taskResults := &schema.Result{
		Field: make([]*schema.Item, 0),
	}
	for name, value := range ep.Results {
		item := &schema.Item{
			Name:  name,
			Value: value.Value,
			Type:  types.ToSchemaType(value.Type),
		}
		taskResults.Field = append(taskResults.Field, item)
	}
	extension.ResultsField = taskResults

	var st schema.TaskInterface
	switch ep.TaskType {
	case types.FlowNodeType_Task:
		st = &schema.Task{}
	case types.FlowNodeType_ReceiveTask:
		st = &schema.ReceiveTask{}
	case types.FlowNodeType_ServiceTask:
		st = &schema.ServiceTask{}
	case types.FlowNodeType_UserTask:
		st = &schema.UserTask{}
	case types.FlowNodeType_ScriptTask:
		st = &schema.ScriptTask{}
	case types.FlowNodeType_ManualTask:
		st = &schema.ManualTask{}
	case types.FlowNodeType_CallActivity:
		st = &schema.CallActivity{}
	case types.FlowNodeType_BusinessRuleTask:
		st = &schema.BusinessRuleTask{}
	default:
		st = &schema.Task{}
	}

	st.SetName(schema.NewStringP(ep.Name))
	st.SetExtensionElements(extension)

	return st
}
