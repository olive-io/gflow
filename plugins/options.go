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

package plugins

import (
	"github.com/olive-io/gflow/api/types"
)

const (
	GflowPlugin string = "gflow"
	GRPCPlugin         = "grpc"
	HTTPPlugin         = "http"

	RabbitMQPlugin = "rabbitmq"
)

// Options contains configuration options for plugin creation.
type Options struct {
	// Type specifies the type for the plugin
	Type string
	// Target specifies the target destination for the plugin
	Target string
}

// NewOptions creates a new Options instance by applying the provided option functions.
func NewOptions(opts ...Option) *Options {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.Type == "" {
		options.Type = GflowPlugin
	}

	return &options
}

// Option is a function type for configuring plugin creation options.
type Option func(o *Options)

// WithType returns an Option function that sets the type for plugin configuration.
func WithType(typ string) Option {
	return func(o *Options) {
		o.Type = typ
	}
}

// WithTarget returns an Option function that sets the target for plugin configuration.
// The target parameter specifies the destination the plugin should use.
func WithTarget(target string) Option {
	return func(o *Options) {
		o.Target = target
	}
}

type RegisterOptions struct {
	Name        string
	FlowType    types.FlowNodeType
	Type        string
	Description string
	Task        any
	Request     any
	Response    any
}

type RegisterOption func(*RegisterOptions)

func NewRegisterOptions(opts ...RegisterOption) *RegisterOptions {
	var options RegisterOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.FlowType == 0 {
		options.FlowType = types.FlowNodeType_ServiceTask
	}
	if options.Type == "" {
		options.Type = GflowPlugin
	}

	return &options
}

func WithTaskName(name string) RegisterOption {
	return func(o *RegisterOptions) {
		o.Name = name
	}
}

func WithFlowType(t types.FlowNodeType) RegisterOption {
	return func(o *RegisterOptions) {
		o.FlowType = t
	}
}

func WithTaskType(typ string) RegisterOption {
	return func(o *RegisterOptions) {
		o.Type = typ
	}
}

func WithDesc(desc string) RegisterOption {
	return func(o *RegisterOptions) {
		o.Description = desc
	}
}

func WithTask(task any) RegisterOption {
	return func(o *RegisterOptions) {
		o.Task = task
	}
}

func WithRequest(request any) RegisterOption {
	return func(o *RegisterOptions) {
		o.Request = request
	}
}

func WithResponse(response any) RegisterOption {
	return func(o *RegisterOptions) {
		o.Response = response
	}
}

// DoOptions contains runtime options for plugin execution.
// It provides metadata storage for passing additional context during plugin execution.
type DoOptions struct {
	Stage   types.CallTaskStage
	Process int64
	Name    string
	Timeout int64
	// metadata stores key-value pairs of execution context
	metadata map[string]any
}

func NewDoOptions(opts ...DoOption) *DoOptions {
	var options DoOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.metadata == nil {
		options.metadata = make(map[string]any)
	}

	return &options
}

// Metadata returns the complete metadata map containing all key-value pairs.
// This provides access to all execution context data stored in the options.
func (do *DoOptions) Metadata() map[string]any {
	return do.metadata
}

// Get retrieves a specific metadata value by key.
// It returns the value and a boolean indicating whether the key exists.
func (do *DoOptions) Get(key string) (any, bool) {
	value, ok := do.metadata[key]
	return value, ok
}

// DoOption is a function type for configuring plugin execution options.
type DoOption func(o *DoOptions)

func DoWithProcess(process int64) DoOption {
	return func(o *DoOptions) {
		o.Process = process
	}
}

func DoWithName(name string) DoOption {
	return func(o *DoOptions) {
		o.Name = name
	}
}

func DoWithTaskStage(stage types.CallTaskStage) DoOption {
	return func(o *DoOptions) {
		o.Stage = stage
	}
}

func DoWithTimeout(timeout int64) DoOption {
	return func(o *DoOptions) {
		o.Timeout = timeout
	}
}

// WithMetadatas returns a DoOption function that sets the entire metadata map.
// This replaces any existing metadata with the provided map.
func WithMetadatas(metadata map[string]any) DoOption {
	return func(o *DoOptions) {
		o.metadata = metadata
	}
}

// WithKV returns a DoOption function that sets a single key-value pair in metadata.
// If the metadata map doesn't exist, it creates a new one.
// This allows for incremental addition of metadata entries.
func WithKV(key string, value any) DoOption {
	return func(o *DoOptions) {
		if o.metadata == nil {
			o.metadata = make(map[string]any)
		}
		o.metadata[key] = value
	}
}
