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
	"context"
	"fmt"
	"reflect"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/plugins"
	"github.com/olive-io/gflow/server/config"
	"github.com/olive-io/gflow/server/plugin/sdk"
)

var _ plugins.Plugin = (*rabbitmq)(nil)

type rabbitmq struct {
	cfg *config.RabbitMQConfig
}

func NewRabbitMQ(cfg *config.RabbitMQConfig) plugins.Plugin {
	p := &rabbitmq{
		cfg: cfg,
	}

	return p
}

func (mq *rabbitmq) Name() string { return plugins.RabbitMQPlugin }

func (mq *rabbitmq) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	var options plugins.DoOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.Stage > types.CallTaskStage_Commit {
		return &plugins.Response{}, nil
	}

	request := new(RabbitRequest)
	if err := req.ApplyTo(request); err != nil {
		return nil, fmt.Errorf("binding requet: %v", err)
	}

	client, err := sdk.NewRabbitmq(mq.cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %v", err)
	}
	defer client.Close()

	if err = client.Send(ctx, request.Topic, request.ContentType, []byte(request.Body)); err != nil {
		return nil, fmt.Errorf("send: %v", err)
	}

	response := &RabbitResponse{Result: true}
	return plugins.ExtractResponse(reflect.ValueOf(response)), nil
}

type RabbitRequest struct {
	ContentType string `gflow:"hr:Content-Type"`
	Topic       string `json:"topic"`
	Body        string `json:"body"`
}

type RabbitResponse struct {
	Result bool `json:"result"`
}
