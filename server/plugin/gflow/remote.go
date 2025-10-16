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
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/dispatch"
)

var _ plugins.Factory = (*remoteFactory)(nil)

type remoteFactory struct {
	dispatcher *dispatch.Dispatcher
}

func NewFactory(dispatcher *dispatch.Dispatcher) (plugins.Factory, error) {
	factory := &remoteFactory{
		dispatcher: dispatcher,
	}
	return factory, nil
}

func (f *remoteFactory) Name() string { return "gflow" }

func (f *remoteFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
	options := plugins.NewOptions(opts...)

	target := options.Target
	if target == "" {
		return nil, fmt.Errorf("%w: target is required", plugins.ErrInvalidCreationOptions)
	}

	gp := &remotePlugin{
		dispatcher: f.dispatcher,
		target:     target,
	}

	return gp, nil
}

var _ plugins.Plugin = (*remotePlugin)(nil)

type remotePlugin struct {
	dispatcher *dispatch.Dispatcher
	target     string
}

func (gp *remotePlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	var doOptions plugins.DoOptions
	for _, opt := range opts {
		opt(&doOptions)
	}

	if doOptions.Name == "" {
		return nil, fmt.Errorf("%w: name is required", plugins.ErrInvalidDoOptions)
	}
	if doOptions.Stage > types.CallTaskStage_Destroy {
		return nil, fmt.Errorf("%w: bad stage '%s'", plugins.ErrInvalidDoOptions, doOptions.Stage.String())
	}

	callRequest := &types.CallTaskRequest{
		Stage:       doOptions.Stage,
		Process:     doOptions.Process,
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
		return nil, fmt.Errorf("%w: %v", plugins.ErrDoExecution, err)
	}

	callResp := &plugins.Response{
		Results:     resp.Results,
		DataObjects: resp.DataObjects,
		Error:       resp.Error,
	}
	return callResp, nil
}

func (gp *remotePlugin) String() string { return "gflow" }
