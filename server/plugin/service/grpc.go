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
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/plugin/sdk"
)

var _ plugins.Plugin = (*grpcPlugin)(nil)

type grpcPlugin struct {
	lg       *zap.Logger
	flowType types.FlowNodeType
	typ      string
}

func newGRPCPlugin(lg *zap.Logger, flowType types.FlowNodeType) *grpcPlugin {
	hp := &grpcPlugin{
		lg:       lg,
		flowType: flowType,
		typ:      plugins.HTTPPlugin,
	}

	return hp
}

func (hp *grpcPlugin) Name() string { return hp.typ }

func (hp *grpcPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	lg := hp.lg
	options := plugins.NewDoOptions(opts...)

	if options.Stage > types.CallTaskStage_Commit {
		return &plugins.Response{}, nil
	}

	gc := &sdk.GRPCConfig{
		Timeout: time.Duration(options.Timeout) * time.Millisecond,
	}
	tv, found := req.Properties["host"]
	if !found {
		return nil, fmt.Errorf("missing 'host' property")
	}
	gc.Host = tv.Value

	grpcConn, err := sdk.NewGRPCClient(lg, gc)
	if err != nil {
		return nil, fmt.Errorf("connect to grpc server: %w", err)
	}

	name := options.Name
	headers := req.Headers
	in := map[string]any{}

	for key, pt := range req.Properties {
		in[key] = pt.ToSchemaValue().Value()
	}

	presp := &plugins.Response{
		Results: make(map[string]*types.Value),
	}

	out := map[string]any{}
	status, err := grpcConn.Call(ctx, name, headers, in, &out)
	if err != nil {
		presp.Error = err.Error()
		return nil, fmt.Errorf("call grpc: %w", err)
	}

	if status.Code() != codes.OK {
		presp.Error = fmt.Sprintf("%s", status.Message())
		return nil, fmt.Errorf("call %s: %w", name, status.Err())
	}

	for k, v := range out {
		presp.Results[k] = types.NewValue(v)
	}

	return presp, nil
}
