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

package gflow

import (
	"context"
	"fmt"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/server/dispatch"
	"github.com/olive-io/gflow/server/plugin"
)

var _ plugin.Factory = (*gflowFactory)(nil)

type gflowFactory struct {
	dispatcher *dispatch.Dispatcher
}

func NewFactory(dispatcher *dispatch.Dispatcher) (plugin.Factory, error) {
	factory := &gflowFactory{
		dispatcher: dispatcher,
	}
	return factory, nil
}

func (f *gflowFactory) Name() string { return "gflow" }

func (f *gflowFactory) Create(opts ...plugin.Option) (plugin.Plugin, error) {
	options := plugin.NewOptions(opts...)

	target := options.Target
	if target == "" {
		return nil, fmt.Errorf("%w: target is required", plugin.ErrInvalidCreationOptions)
	}

	gp := &gflowPlugin{
		dispatcher: f.dispatcher,
		target:     target,
	}

	return gp, nil
}

var _ plugin.Plugin = (*gflowPlugin)(nil)

type gflowPlugin struct {
	dispatcher *dispatch.Dispatcher
	target     string
}

func (gp *gflowPlugin) Do(ctx context.Context, req *plugin.Request, opts ...plugin.DoOption) (*plugin.Response, error) {
	var doOptions plugin.DoOptions
	for _, opt := range opts {
		opt(&doOptions)
	}

	callRequest := &types.CallTaskRequest{
		Stage:       doOptions.Stage,
		FlowType:    doOptions.FlowType,
		Kind:        doOptions.Kind,
		Name:        doOptions.Name,
		Headers:     req.Headers,
		Properties:  req.Properties,
		DataObjects: req.DataObjects,
		Timeout:     doOptions.Timeout,
	}

	callOptions := make([]dispatch.CallOption, 0)
	callOptions = append(callOptions, dispatch.WithUID(gp.target))

	resp, err := gp.dispatcher.CallTask(ctx, callRequest, callOptions...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", plugin.ErrDoExecution, err)
	}

	callResp := &plugin.Response{
		Results:     resp.Results,
		DataObjects: resp.DataObjects,
		Error:       resp.Error,
	}
	return callResp, nil
}

func (gp *gflowPlugin) String() string { return "gflow" }
