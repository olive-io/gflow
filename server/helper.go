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
	"fmt"
	urlpkg "net/url"
	"path"
	"strings"

	json "github.com/bytedance/sonic"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/olive-io/bpmn/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/yaml"

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
	case types.FlowNodeType_SendTask:
		st = &schema.SendTask{}
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

func parseDoc(doc, ct string) (*types.Runner, []*types.Endpoint, error) {
	match, err := isSwaggerDoc(doc, ct)
	if err != nil {
		return nil, nil, err
	}
	if match {
		return parseSwaggerDoc(doc, ct)
	}

	match, err = isOpenAPIDoc(doc, ct)
	if err != nil {
		return nil, nil, err
	}
	if match {
		return parseOpenAPIDoc(doc, ct)
	}

	return nil, nil, fmt.Errorf("invalid doc")
}

func parseSwaggerDoc(doc, ct string) (*types.Runner, []*types.Endpoint, error) {
	document := openapi2.T{}

	switch ct {
	case "json":
		if err := json.UnmarshalString(doc, &document); err != nil {
			return nil, nil, err
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal([]byte(doc), &document); err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("invalid content type: %s", ct)
	}

	info := document.Info

	runner := &types.Runner{
		Version:     info.Version,
		HeartbeatMs: 30000,
		Metadata:    map[string]string{},
		Features:    map[string]string{},
		Transport:   types.Runner_HTTP,
		State:       types.Runner_Ok,
	}
	host := document.Host
	if document.BasePath != "" {
		host = path.Join(host, document.BasePath)
	}
	if len(document.Schemes) != 0 {
		host = document.Schemes[0] + "://" + host
	}
	runner.ListenUrl = host
	runner.Uid = host
	listenURL, err := urlpkg.Parse(host)
	if err == nil {
		runner.Uid = listenURL.Host
	}

	endpoints := make([]*types.Endpoint, 0)
	for url, pathItem := range document.Paths {
		var operation *openapi2.Operation
		var method string
		paths := map[string]*openapi2.Operation{}
		if pathItem.Get != nil {
			operation = pathItem.Get
			method = "GET"
			paths[method] = operation
		}
		if pathItem.Post != nil {
			operation = pathItem.Post
			method = "POST"
			paths[method] = operation
		}
		if pathItem.Put != nil {
			operation = pathItem.Put
			method = "PUT"
			paths[method] = operation
		}
		if pathItem.Patch != nil {
			operation = pathItem.Patch
			method = "PATCH"
			paths[method] = operation
		}
		if pathItem.Delete != nil {
			operation = pathItem.Delete
			method = "DELETE"
			paths[method] = operation
		}

		for method, operation = range paths {

			pathURL := host + url
			endpoint := &types.Endpoint{
				TaskType:    types.FlowNodeType_ServiceTask,
				Type:        "http",
				Name:        operation.OperationID,
				Description: operation.Description,
				Mode:        types.TransitionMode_Simple,
				HttpUrl:     pathURL,
				Metadata: map[string]string{
					"name":    operation.OperationID,
					"method":  method,
					"url":     pathURL,
					"swagger": document.Swagger,
				},
				Targets:     []string{runner.Uid},
				Headers:     map[string]string{},
				Properties:  map[string]*types.Value{},
				DataObjects: map[string]*types.Value{},
				Results:     map[string]*types.Value{},
			}

			for _, parameter := range operation.Parameters {
				var tv *types.Value
				switch parameter.In {
				case "path":
					tv = openapi2ParameterToValue(parameter)
				case "query":
				case "body":
					schemaRef := parameter.Schema
					if schemaRef.Ref != "" {
						ref := strings.TrimPrefix(schemaRef.Ref, "#/definitions/")
						schemaRef = document.Definitions[ref]
					}
					tv = openapi2SchemaToValue(schemaRef.Value)
				}
				if tv != nil {
					endpoint.Properties[parameter.Name] = tv
				}
			}

			if responseRef := operation.Responses["200"]; responseRef != nil {
				if rv := responseRef.Schema; rv != nil {
					schemaRef := rv
					if ref := schemaRef.Ref; ref != "" {
						ref = strings.TrimPrefix(ref, "#/definitions/")
						schemaRef = document.Definitions[ref]
					}

					properties := openapi2SchemaRefToProperties(schemaRef, &document)
					for name, value := range properties {
						endpoint.Results[name] = value
					}
				}
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	return runner, endpoints, nil
}

func parseOpenAPIDoc(doc, ct string) (*types.Runner, []*types.Endpoint, error) {
	document := openapi3.T{}

	switch ct {
	case "json":
		if err := json.UnmarshalString(doc, &document); err != nil {
			return nil, nil, err
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal([]byte(doc), &document); err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, fmt.Errorf("invalid content type: %s", ct)
	}

	info := document.Info

	runner := &types.Runner{
		Version:     info.Version,
		HeartbeatMs: 30000,
		Metadata:    map[string]string{},
		Features:    map[string]string{},
		Transport:   types.Runner_HTTP,
		State:       types.Runner_Ok,
	}
	if len(document.Servers) > 0 {
		host := document.Servers[0].URL
		runner.ListenUrl = host
		runner.Uid = host
		url, err := urlpkg.Parse(host)
		if err == nil {
			runner.Uid = url.Host
		}
	}

	endpoints := make([]*types.Endpoint, 0)
	for url, pathItem := range document.Paths.Map() {
		var operation *openapi3.Operation
		var method string
		paths := map[string]*openapi3.Operation{}
		if pathItem.Get != nil {
			operation = pathItem.Get
			method = "GET"
			paths[method] = operation
		}
		if pathItem.Post != nil {
			operation = pathItem.Post
			method = "POST"
			paths[method] = operation
		}
		if pathItem.Put != nil {
			operation = pathItem.Put
			method = "PUT"
			paths[method] = operation
		}
		if pathItem.Patch != nil {
			operation = pathItem.Patch
			method = "PATCH"
			paths[method] = operation
		}
		if pathItem.Delete != nil {
			operation = pathItem.Delete
			method = "DELETE"
			paths[method] = operation
		}

		for method, operation = range paths {
			pathURL := path.Join(runner.ListenUrl, url)
			endpoint := &types.Endpoint{
				TaskType:    types.FlowNodeType_ServiceTask,
				Type:        "http",
				Name:        operation.OperationID,
				Description: operation.Description,
				Mode:        types.TransitionMode_Simple,
				HttpUrl:     pathURL,
				Metadata: map[string]string{
					"name":    operation.OperationID,
					"method":  method,
					"url":     pathURL,
					"openapi": document.OpenAPI,
				},
				Targets:     []string{runner.Uid},
				Headers:     map[string]string{},
				Properties:  map[string]*types.Value{},
				DataObjects: map[string]*types.Value{},
				Results:     map[string]*types.Value{},
			}

			for _, parameter := range operation.Parameters {
				pv := parameter.Value
				if pv == nil {
					continue
				}

				schemaRef := pv.Schema
				tv := openapi3SchemaToValue(schemaRef.Value)
				endpoint.Properties[pv.Name] = tv
			}

			if rb := operation.RequestBody; rb != nil {
				if rv := rb.Value; rv != nil {
					mt := rv.Content.Get("application/json")
					if mt != nil {
						schemaRef := mt.Schema
						if ref := schemaRef.Ref; ref != "" {
							ref = strings.TrimPrefix(ref, "#/components/schemas/")
							schemaRef = document.Components.Schemas[ref]
						}

						properties := openapi3SchemaRefToProperties(schemaRef, &document)
						for name, value := range properties {
							endpoint.Properties[name] = value
						}
					}
				}
			}

			if responseRef := operation.Responses.Value("200"); responseRef != nil {
				if rv := responseRef.Value; rv != nil {
					schemaRef := rv.Content.Get("application/json").Schema
					if ref := schemaRef.Ref; ref != "" {
						ref = strings.TrimPrefix(ref, "#/components/schemas/")
						schemaRef = document.Components.Schemas[ref]
					}

					properties := openapi3SchemaRefToProperties(schemaRef, &document)
					for name, value := range properties {
						endpoint.Results[name] = value
					}
				}
			}

			endpoints = append(endpoints, endpoint)
		}
	}

	return runner, endpoints, nil
}

func openapi2SchemaRefToProperties(ref *openapi2.SchemaRef, t *openapi2.T) map[string]*types.Value {
	properties := make(map[string]*types.Value)
	value := ref.Value
	if ref.Ref != "" {
		name := strings.TrimPrefix(ref.Ref, "#/definitions/")
		ref = t.Definitions[name]
		return openapi2SchemaRefToProperties(ref, t)
	}

	for name, pt := range value.Properties {
		pv := pt.Value
		if pv == nil {
			name = strings.TrimPrefix(ref.Ref, "#/definitions/")
			var ok bool
			ref, ok = t.Definitions[name]
			if !ok {
				continue
			}
			pv = ref.Value
		}

		tv := openapi2SchemaToValue(pv)
		properties[name] = tv
	}

	return properties
}

func openapi2SchemaToValue(schemaV2 *openapi2.Schema) *types.Value {
	tv := &types.Value{}

	st := schemaV2.Type
	if st.Is("string") {
		tv.Type = types.Value_String
	} else if st.Is("integer") {
		tv.Type = types.Value_Integer
	} else if st.Is("number") {
		tv.Type = types.Value_Float
	} else if st.Is("boolean") {
		tv.Type = types.Value_Boolean
	} else if st.Is("array") {
		tv.Type = types.Value_Array
	} else if st.Is("object") {
		tv.Type = types.Value_Object
	} else {
		tv.Type = types.Value_Object
	}

	return tv
}

func openapi2ParameterToValue(p *openapi2.Parameter) *types.Value {
	tv := &types.Value{}

	st := p.Type
	if st.Is("string") {
		tv.Type = types.Value_String
	} else if st.Is("integer") {
		tv.Type = types.Value_Integer
	} else if st.Is("number") {
		tv.Type = types.Value_Float
	} else if st.Is("boolean") {
		tv.Type = types.Value_Boolean
	} else if st.Is("array") {
		tv.Type = types.Value_Array
	} else if st.Is("object") {
		tv.Type = types.Value_Object
	} else {
		tv.Type = types.Value_Object
	}

	return tv
}

func openapi3SchemaRefToProperties(ref *openapi3.SchemaRef, t *openapi3.T) map[string]*types.Value {
	properties := make(map[string]*types.Value)
	value := ref.Value
	if ref.Ref != "" {
		name := strings.TrimPrefix(ref.Ref, "#/components/schemas/")
		ref = t.Components.Schemas[name]
		value = ref.Value
	}

	for name, pt := range value.Properties {
		pv := pt.Value
		if pv == nil {
			name = strings.TrimPrefix(pt.Ref, "#/components/schemas/")
			var ok bool
			ref, ok = t.Components.Schemas[name]
			if !ok {
				continue
			}
			pv = ref.Value
		}

		tv := openapi3SchemaToValue(pv)
		properties[name] = tv
	}

	return properties
}

func openapi3SchemaToValue(schemaV3 *openapi3.Schema) *types.Value {
	tv := &types.Value{}

	st := schemaV3.Type
	if st.Is("string") {
		tv.Type = types.Value_String
	} else if st.Is("integer") {
		tv.Type = types.Value_Integer
	} else if st.Is("number") {
		tv.Type = types.Value_Float
	} else if st.Is("boolean") {
		tv.Type = types.Value_Boolean
	} else if st.Is("array") {
		tv.Type = types.Value_Array
	} else if st.Is("object") {
		tv.Type = types.Value_Object
	} else {
		tv.Type = types.Value_Object
	}

	return tv
}

func isSwaggerDoc(doc, ct string) (bool, error) {
	metadata := map[string]any{}

	switch ct {
	case "json":
		if err := json.UnmarshalString(doc, &metadata); err != nil {
			return false, err
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal([]byte(doc), &metadata); err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("invalid content type: %s", ct)
	}

	_, ok := metadata["swagger"]
	return ok, nil
}

func isOpenAPIDoc(doc, ct string) (bool, error) {
	metadata := map[string]any{}

	switch ct {
	case "json":
		if err := json.UnmarshalString(doc, &metadata); err != nil {
			return false, err
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal([]byte(doc), &metadata); err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("invalid content type: %s", ct)
	}

	_, ok := metadata["openapi"]
	return ok, nil
}

func getPageSize(in any) (int, int) {
	impl, ok := in.(interface {
		GetPage() int32
		GetSize() int32
	})
	if !ok {
		return 1, 10
	}

	page, size := int(impl.GetPage()), int(impl.GetSize())
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 10
	}
	return page, size
}
