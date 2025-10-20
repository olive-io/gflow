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

package send

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/config"
)

var _ plugins.Factory = (*sendFactory)(nil)

type sendFactory struct {
	*plugins.TaskFactoryImpl
	lg       *zap.Logger
	taskType types.FlowNodeType

	cfg *config.SendTaskPluginConfig
}

func NewFactory(lg *zap.Logger, cfg *config.SendTaskPluginConfig) (plugins.Factory, error) {
	taskType := types.FlowNodeType_SendTask
	impl := plugins.NewTaskFactory(taskType.String())
	factory := &sendFactory{
		TaskFactoryImpl: impl,
		lg:              lg,
		taskType:        taskType,
		cfg:             cfg,
	}

	if err := factory.Register(
		plugins.WithTaskName("rabbitmq.send"),
		plugins.WithTaskType("rabbitmq"),
		plugins.WithFlowType(types.FlowNodeType_SendTask),
		plugins.WithRequest(new(RabbitRequest)),
		plugins.WithResponse(new(RabbitResponse))); err != nil {
		return nil, fmt.Errorf("failed to register rabbitmq plugin: %w", err)
	}

	return factory, nil
}

func (f *sendFactory) Name() string { return f.taskType.String() }

func (f *sendFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
	lg := f.lg
	options := plugins.NewOptions(opts...)

	lg.Debug("create plugins",
		zap.String("taskType", f.taskType.String()),
		zap.String("type", options.Type))

	var rp plugins.Plugin
	switch options.Type {
	case plugins.RabbitMQPlugin:
		mqConfig := f.cfg.RabbitMQ
		if mqConfig == nil {
			return nil, fmt.Errorf("missing rabbitmq config")
		}

		rp = NewRabbitMQ(mqConfig.RabbitMQConfig)
	default:
		return nil, fmt.Errorf("invalid plugin type: %s", options.Type)
	}

	return rp, nil
}
