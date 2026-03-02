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

package script

import (
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/config"
)

var _ plugins.Factory = (*scriptFactory)(nil)

type scriptFactory struct {
	lg       *otelzap.Logger
	taskType types.FlowNodeType
	cfg      *config.ScriptTaskPluginConfig
}

func NewFactory(lg *otelzap.Logger, cfg *config.ScriptTaskPluginConfig) (plugins.Factory, error) {
	taskType := types.FlowNodeType_ScriptTask
	factory := &scriptFactory{
		lg:       lg,
		taskType: taskType,
		cfg:      cfg,
	}

	return factory, nil
}

func (f *scriptFactory) Name() string { return f.taskType.String() }

func (f *scriptFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
	lg := f.lg
	options := plugins.NewOptions(opts...)

	lg.Debug("create script plugin",
		zap.String("taskType", f.taskType.String()),
		zap.String("type", options.Type))

	var rp plugins.Plugin
	switch options.Type {
	case ShellPlugin:
		rp = NewShellPlugin(f.cfg)
	case PythonPlugin:
		rp = NewPythonPlugin(f.cfg)
	case "":
		rp = NewShellPlugin(f.cfg)
	default:
		return nil, fmt.Errorf("invalid script plugin type: %s", options.Type)
	}

	return rp, nil
}
