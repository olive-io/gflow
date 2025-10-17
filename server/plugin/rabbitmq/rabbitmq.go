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

package rabbitmq

import (
	"context"

	"github.com/olive-io/gflow/plugins"
)

var _ plugins.Factory = (*rabbitmqFactory)(nil)

type rabbitmqFactory struct{}

func (rf *rabbitmqFactory) Name() string { return "rabbitmq" }

func (rf *rabbitmqFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
	rp := &rabbitmqPlugin{}

	return rp, nil
}

type rabbitmqPlugin struct{}

func (rp *rabbitmqPlugin) Name() string { return "rabbitmq" }

func (rp *rabbitmqPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	//TODO implement me
	panic("implement me")
}
