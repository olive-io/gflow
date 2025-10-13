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

package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
)

type Task interface {
	Commit(ctx context.Context, request any) (any, error)
	Rollback(ctx context.Context) error
	Destroy(ctx context.Context) error
	String() string
}

var _ Task = (*TaskImpl)(nil)

type TaskImpl struct {
	name        string
	headers     map[string]string
	properties  map[string]*types.Value
	dataObjects map[string]*types.Value

	results map[string]*types.Value
}

func (impl *TaskImpl) Commit(ctx context.Context, request any) (any, error) {
	resp := &struct{}{}
	return resp, nil
}

func (impl *TaskImpl) Rollback(ctx context.Context) error {
	return nil
}

func (impl *TaskImpl) Destroy(ctx context.Context) error {
	return nil
}

func (impl *TaskImpl) String() string { return impl.name }

type Controller struct {
	tasks map[string]Task

	mu   sync.RWMutex
	pool map[string]Task
}

func NewController() *Controller {
	controller := &Controller{
		tasks: make(map[string]Task),
		pool:  make(map[string]Task),
	}
	return controller
}

func (c *Controller) addTaskDefine(task Task) {}

func (c *Controller) ListEndpoints() []*types.Endpoint {
	endpoints := make([]*types.Endpoint, 0)
	return endpoints
}

type Options struct {
	Name     string
	Request  any
	Response any
}

type Option func(*Options)

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithRequest(request any) Option {
	return func(o *Options) {
		o.Request = request
	}
}

func WithResponse(response any) Option {
	return func(o *Options) {
		o.Response = response
	}
}

func (r *Runner) Register(task Task, opts ...Option) error { return nil }

func (r *Runner) RegisterFn(fn any, opts ...Option) error { return nil }

func (r *Runner) handle(ctx context.Context, req *types.CallTaskRequest) (*types.CallTaskResponse, error) {
	lg := r.lg
	timeout := time.Duration(req.Timeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	results := make(map[string]*types.Value)
	results["a"] = types.NewValue("bb")

	resp := &types.CallTaskResponse{
		SeqId:   req.SeqId,
		Results: results,
	}

	lg.Info("call response", zap.Any("request", req))

	return resp, fmt.Errorf("internal error")
}
