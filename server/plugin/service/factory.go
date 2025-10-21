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
	"fmt"

	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/dispatch"
)

var _ plugins.Factory = (*remoteFactory)(nil)

type remoteFactory struct {
	lg         *zap.Logger
	taskType   types.FlowNodeType
	dispatcher *dispatch.Dispatcher
}

func NewFactory(lg *zap.Logger, dispatcher *dispatch.Dispatcher) (plugins.Factory, error) {
	factory := &remoteFactory{
		lg:         lg,
		taskType:   types.FlowNodeType_ServiceTask,
		dispatcher: dispatcher,
	}
	return factory, nil
}

func (f *remoteFactory) Name() string { return f.taskType.String() }

func (f *remoteFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
	lg := f.lg
	options := plugins.NewOptions(opts...)

	target := options.Target
	if target == "" {
		return nil, fmt.Errorf("%w: target is required", plugins.ErrInvalidCreationOptions)
	}

	var gp plugins.Plugin
	switch options.Type {
	case plugins.GflowPlugin:
		gp = newRemoteGflow(lg, f.dispatcher, f.taskType, target)
	case plugins.GRPCPlugin:
		gp = newHTTPPlugin(lg, f.taskType)
	case plugins.HTTPPlugin:
		gp = newGRPCPlugin(lg, f.taskType)
	default:
		return nil, fmt.Errorf("invalid plugin type: %s", options.Type)
	}

	return gp, nil
}
