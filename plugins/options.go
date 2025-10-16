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

// Options contains configuration options for plugin creation.
// It includes target configuration and can be extended with additional plugin-specific options.
type Options struct {
	// Target specifies the target destination or endpoint for the plugin
	Target string
}

// NewOptions creates a new Options instance by applying the provided option functions.
func NewOptions(opts ...Option) *Options {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	return &options
}

// Option is a function type for configuring plugin creation options.
type Option func(o *Options)

// WithTarget returns an Option function that sets the target for plugin configuration.
// The target parameter specifies the destination or endpoint the plugin should use.
func WithTarget(target string) Option {
	return func(o *Options) {
		o.Target = target
	}
}

// DoOptions contains runtime options for plugin execution.
// It provides metadata storage for passing additional context during plugin execution.
type DoOptions struct {
	Stage    types.CallTaskStage
	Process  int64
	FlowType types.FlowNodeType
	Kind     string
	Name     string
	Timeout  int64
	// metadata stores key-value pairs of execution context
	metadata map[string]any
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

func DoWithKind(kind string) DoOption {
	return func(o *DoOptions) {
		o.Kind = kind
	}
}

func DoWithTaskType(flowType types.FlowNodeType) DoOption {
	return func(o *DoOptions) {
		o.FlowType = flowType
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
