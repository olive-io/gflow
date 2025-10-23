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
	"reflect"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/olive-io/gflow/api/types"
	"github.com/olive-io/gflow/pkg/inject"
	"github.com/olive-io/gflow/plugins"
)

var _ plugins.Factory = (*TaskFactory)(nil)

type TaskFactory struct {
	taskType types.FlowNodeType

	endpoints  map[string]*types.Endpoint
	prototypes map[string]Task

	cmu       sync.RWMutex
	cachePool map[string]Task
}

func NewFactory(taskType types.FlowNodeType) *TaskFactory {
	tf := &TaskFactory{
		taskType:   taskType,
		endpoints:  make(map[string]*types.Endpoint),
		prototypes: make(map[string]Task),
		cachePool:  make(map[string]Task),
	}

	return tf
}

func (tf *TaskFactory) ListEndpoint() []types.Endpoint {
	endpoints := make([]types.Endpoint, 0, len(tf.endpoints))
	for _, ep := range tf.endpoints {
		endpoints = append(endpoints, *ep.DeepCopy())
	}
	return endpoints
}

func (tf *TaskFactory) Register(opts ...plugins.RegisterOption) error {
	options := plugins.NewRegisterOptions(opts...)

	target := options.Task
	if target == nil {
		return fmt.Errorf("missing task")
	}

	var endpoint *types.Endpoint
	var proxy Task
	var err error
	if task, ok := target.(Task); ok {
		endpoint, proxy, err = extractTask(task, options)
		if err != nil {
			return fmt.Errorf("register task: %w", err)
		}
	} else {
		rt := reflect.TypeOf(target)
		if rt.Kind() != reflect.Ptr {
			rt = rt.Elem()
		}
		switch rt.Kind() {
		case reflect.Func:
			endpoint, proxy, err = extractFunc(target, options)
			if err != nil {
				return fmt.Errorf("register function: %w", err)
			}
		default:
			return fmt.Errorf("register target '%s' is not a function", rt.String())
		}
	}

	name := endpoint.Name
	_, exists := tf.endpoints[name]
	if exists {
		return fmt.Errorf("endpoint '%s' already exists", name)
	}
	tf.endpoints[name] = endpoint

	tf.prototypes[name] = proxy
	return nil
}

func (tf *TaskFactory) createPrototype(options *plugins.DoOptions) (Task, func(), error) {
	name := options.Name

	cancel := func() {}
	pt, found := tf.prototypes[name]
	if !found {
		return nil, cancel, fmt.Errorf("task '%s' not found", name)
	}

	ident := fmt.Sprintf("%d.%s", options.Process, options.Name)
	switch options.Stage {
	case types.CallTaskStage_Echo, types.CallTaskStage_Commit:
		task := pt
		if impl, ok := task.(TaskClone); ok {
			task = impl.Clone().(Task)
		}

		if options.Stage == types.CallTaskStage_Commit {
			cancel = func() {
				tf.cmu.Lock()
				tf.cachePool[ident] = task
				tf.cmu.Unlock()
			}
		}

		return task, cancel, nil

	case types.CallTaskStage_Rollback:
		tf.cmu.RLock()
		cacheTask, ok := tf.cachePool[ident]
		tf.cmu.RUnlock()
		if !ok {
			return nil, cancel, fmt.Errorf("rollback task '%s' not found", name)
		}

		return cacheTask, cancel, nil

	case types.CallTaskStage_Destroy:
		tf.cmu.RLock()
		cacheTask, ok := tf.cachePool[ident]
		tf.cmu.RUnlock()

		if !ok {
			return nil, cancel, fmt.Errorf("destroy task '%s' not found", name)
		}

		cancel = func() {
			tf.cmu.Lock()
			delete(tf.cachePool, ident)
			tf.cmu.Unlock()
		}

		return cacheTask, cancel, nil
	default:
		return nil, cancel, fmt.Errorf("unknown stage '%s'", options.Stage)
	}
}

func (tf *TaskFactory) Name() string { return tf.taskType.String() }

func (tf *TaskFactory) Create(opts ...plugins.Option) (plugins.Plugin, error) {
	options := plugins.NewOptions(opts...)

	var tp *taskPlugin
	switch options.Type {
	case plugins.GflowPlugin:
		tp = &taskPlugin{
			typ:             plugins.GflowPlugin,
			createPrototype: tf.createPrototype,
		}
	default:
		return nil, fmt.Errorf("invalid plugin type: %s", options.Type)
	}

	return tp, nil
}

type taskPlugin struct {
	typ string

	createPrototype func(options *plugins.DoOptions) (Task, func(), error)
}

func (tp *taskPlugin) Name() string { return tp.typ }

func (tp *taskPlugin) Do(ctx context.Context, req *plugins.Request, opts ...plugins.DoOption) (*plugins.Response, error) {
	options := plugins.NewDoOptions(opts...)

	taskImpl, clean, err := tp.createPrototype(options)
	if err != nil {
		return nil, fmt.Errorf("create task '%s': %v", tp.typ, err)
	}
	defer clean()

	taskCounter.Add(1)
	defer taskCounter.Sub(1)

	switch options.Stage {
	case types.CallTaskStage_Echo, types.CallTaskStage_Commit:
		taskCommitCounter.Add(1)
		defer taskCommitCounter.Sub(1)

		if err = inject.PopulateTarget(taskImpl); err != nil {
			return nil, fmt.Errorf("inject task '%s': %v", options.Name, err)
		}

		resp, err := taskImpl.Commit(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp.(*plugins.Response), nil

	case types.CallTaskStage_Rollback:
		taskRollbackCounter.Add(1)
		defer taskRollbackCounter.Sub(1)

		err = taskImpl.Rollback(ctx)
		if err != nil {
			return nil, err
		}
		return &plugins.Response{}, nil

	case types.CallTaskStage_Destroy:
		taskDestroyCounter.Add(1)
		defer taskDestroyCounter.Sub(1)

		err = taskImpl.Destroy(ctx)
		if err != nil {
			return nil, err
		}
		return &plugins.Response{}, nil

	default:
		return nil, fmt.Errorf("unknown stage '%s'", options.Stage)
	}
}

func (r *Runner) SetupFactory(factory plugins.Factory) error {
	return r.pluginManager.Setup(factory)
}

func (r *Runner) Provide(values ...any) error {
	return inject.Provide(values...)
}

func (r *Runner) ListEndpoints() []*types.Endpoint {
	endpoints := make([]*types.Endpoint, 0)
	for _, item := range r.pluginManager.ListEndpoints() {
		endpoints = append(endpoints, item.DeepCopy())
	}
	return endpoints
}

func (r *Runner) Handle(ctx context.Context, request *types.CallTaskRequest) *types.CallTaskResponse {
	lg := r.lg
	timeout := time.Duration(request.Timeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response := &types.CallTaskResponse{
		Stage:       request.Stage,
		SeqId:       request.SeqId,
		Results:     map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
	}

	var err error
	var span trace.Span
	ctx, span = r.Tracer().Start(ctx, request.Name,
		trace.WithSpanKind(trace.SpanKindClient))

	defer func() {
		if err != nil {
			span.RecordError(err)
		}
		span.End()
	}()

	kind := request.TaskType.String()
	var factory plugins.Factory
	factory, err = r.pluginManager.Get(kind)
	if err != nil {
		response.Error = fmt.Sprintf("plugin factory '%s' not found", kind)
		return response
	}

	typ := request.Type
	options := make([]plugins.Option, 0)
	options = append(options, plugins.WithType(typ))

	var plugin plugins.Plugin
	plugin, err = factory.Create(options...)
	if err != nil {
		response.Error = fmt.Sprintf("plugin '%s' not found", typ)
		return response
	}

	req := &plugins.Request{
		Headers:     request.Headers,
		Properties:  request.Properties,
		DataObjects: request.DataObjects,
	}

	doOptions := []plugins.DoOption{
		plugins.DoWithProcess(request.Process),
		plugins.DoWithTaskStage(request.Stage),
		plugins.DoWithName(request.Name),
	}

	lg.Info("call task",
		zap.Int64("process", request.Process),
		zap.String("stage", request.Stage.String()),
		zap.String("task", request.Name))

	ctx = context.WithValue(ctx, originKey, request)

	var resp *plugins.Response
	resp, err = plugin.Do(ctx, req, doOptions...)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Results = resp.Results
	response.DataObjects = resp.DataObjects
	return response
}
