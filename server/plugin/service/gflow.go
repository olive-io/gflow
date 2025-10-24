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

package service

import (
	"context"
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/dispatch"
)

var _ plugins.Plugin = (*remoteGflow)(nil)

type remoteGflow struct {
	lg         *otelzap.Logger
	dispatcher *dispatch.Dispatcher
	taskType   types.FlowNodeType
	typ        string
	target     string
}

func newRemoteGflow(lg *otelzap.Logger, dispatcher *dispatch.Dispatcher, taskType types.FlowNodeType, target string) *remoteGflow {
	rgp := &remoteGflow{
		lg:         lg,
		dispatcher: dispatcher,
		taskType:   taskType,
		typ:        plugins.GflowPlugin,
		target:     target,
	}
	return rgp
}

func (gp *remoteGflow) Name() string { return gp.typ }

func (gp *remoteGflow) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
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
		TaskType:    gp.taskType,
		Type:        gp.typ,
		Name:        doOptions.Name,
		Headers:     req.Headers,
		Properties:  req.Properties,
		DataObjects: req.DataObjects,
		Timeout:     doOptions.Timeout,
	}

	// set trace span config
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	traceSpan := &types.SpanContext{
		TraceId: sc.TraceID().String(),
		SpanId:  sc.SpanID().String(),
		Flags:   int32(sc.TraceFlags()),
		State:   sc.TraceState().String(),
		Remote:  true,
	}
	callRequest.TraceSpanContext = traceSpan

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
