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

// Package plugins provides a plugin management system for gflow.
// It allows dynamic registration and retrieval of plugins at runtime.
package plugins

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/olive-io/gflow/api/types"
)

var (
	// defaultManager is the global plugin manager instance
	defaultManager *Manager
	// once ensures the defaultManager is initialized only once
	once sync.Once
)

var (
	ErrFactoryNotFound        = errors.New("plugin factory not found")
	ErrFactoryAlreadyExists   = errors.New("plugin factory already exists")
	ErrInvalidCreationOptions = errors.New("invalid factory creation options")
	ErrInvalidDoOptions       = errors.New("invalid factory do options")
	ErrDoExecution            = errors.New("plugin executes failed")
)

// init initializes the default plugin manager using sync.Once to ensure thread safety
func init() {
	once.Do(func() {
		defaultManager = NewManager()
	})
}

// Setup adds a plugin factory with the default manager.
// It returns an error if a plugin with the same name is already registered.
func Setup(factory Factory) error {
	return defaultManager.Setup(factory)
}

// Get retrieves a plugin factory by name from the default manager.
// It returns an error if the plugin is not found.
func Get(name string) (Factory, error) {
	return defaultManager.Get(name)
}

// ListEndpoints retrieves all *types.Endpoint from the default manager.
func ListEndpoints() []*types.Endpoint {
	return defaultManager.ListEndpoints()
}

// Request represents a plugin execution request containing headers, properties,
// data objects, and timeout configuration.
type Request struct {
	// Headers contains HTTP-style headers for the request
	Headers map[string]string
	// Properties contains key-value properties for plugin configuration
	Properties map[string]*types.Value
	// DataObjects contains structured data objects for plugin processing
	DataObjects map[string]*types.Value
}

// Response represents the result of plugin execution containing results and data objects.
type Response struct {
	// Results contains the output values from plugin execution
	Results map[string]*types.Value
	// DataObjects contains structured data objects returned by the plugin
	DataObjects map[string]*types.Value
	// Error contains the error message from plugin execution
	Error string
}

// Plugin defines the interface that all plugins must implement.
type Plugin interface {
	// Name returns the unique name of the plugin type
	Name() string
	// Do executes the plugin with the given request and returns a response
	Do(ctx context.Context, req *Request, opts ...DoOption) (*Response, error)
}

// TaskPlugin defines the interface that all plugins must implement and returns types.Endpoint
type TaskPlugin interface {
	Plugin
	// GetEndpoint returns relational Task or function
	GetEndpoint() types.Endpoint
}

// Factory defines the interface for creating plugin instances.
// Each factory is responsible for creating plugins of a specific type.
type Factory interface {
	// Name returns the unique name of the plugin type
	Name() string
	// Create creates a new plugin instance with the given options
	Create(opts ...Option) (Plugin, error)
}

// TaskFactory defines the interface that all factories must implement and returns types.Endpoint
type TaskFactory interface {
	Factory
	Register(opts ...RegisterOption) error
	// ListEndpoint returns relational Tasks or functions
	ListEndpoint() []types.Endpoint
}

// Manager manages plugin factories and provides registration and retrieval functionality.
type Manager struct {
	// factories stores registered plugin factories indexed by name
	factories map[string]Factory
	// endpoints stores registered plugin Endpoint indexed by name
	endpoints map[string]types.Endpoint
}

// NewManager creates a new plugin manager with an empty factory registry.
func NewManager() *Manager {
	manager := &Manager{
		factories: make(map[string]Factory),
		endpoints: make(map[string]types.Endpoint),
	}

	return manager
}

// Setup adds a plugin factory with the manager.
// It returns an error if a factory with the same name is already registered.
func (m *Manager) Setup(factory Factory) error {
	name := factory.Name()
	_, ok := m.factories[name]
	if ok {
		return fmt.Errorf("%w: %s", ErrFactoryAlreadyExists, name)
	}

	if v, match := factory.(TaskFactory); match {
		for _, endpoint := range v.ListEndpoint() {
			endpointName := endpoint.Name
			_, exists := m.endpoints[endpointName]
			if exists {
				return fmt.Errorf("%w: endpoint '%s' be registered", ErrFactoryAlreadyExists, endpointName)
			}
			m.endpoints[endpointName] = endpoint
		}
	}

	m.factories[name] = factory
	return nil
}

// Get retrieves a plugin factory by name from the manager.
// It returns an error if no factory with the given name is found.
func (m *Manager) Get(name string) (Factory, error) {
	factory, ok := m.factories[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFactoryNotFound, name)
	}
	return factory, nil
}

// ListEndpoints retrieves all *types.Endpoint from the manager.
func (m *Manager) ListEndpoints() []*types.Endpoint {
	endpoints := make([]*types.Endpoint, 0, len(m.endpoints))
	for _, endpoint := range m.endpoints {
		endpoints = append(endpoints, endpoint.DeepCopy())
	}
	return endpoints
}

var _ TaskFactory = (*TaskFactoryImpl)(nil)

type TaskFactoryImpl struct {
	name      string
	endpoints map[string]*types.Endpoint
}

func NewTaskFactory(name string) *TaskFactoryImpl {
	f := &TaskFactoryImpl{
		endpoints: make(map[string]*types.Endpoint),
	}
	return f
}

func (f *TaskFactoryImpl) Name() string { return f.name }

func (f *TaskFactoryImpl) Create(opts ...Option) (Plugin, error) {
	return nil, fmt.Errorf("implement me")
}

func (f *TaskFactoryImpl) Register(opts ...RegisterOption) error {
	options := NewRegisterOptions(opts...)

	if options.Request == nil {
		return fmt.Errorf("no request provided")
	}
	if options.Response == nil {
		return fmt.Errorf("no response provided")
	}
	if options.Name == "" {
		return fmt.Errorf("missing plugin name")
	}

	endpoint := &types.Endpoint{
		TaskType:    options.FlowType,
		Type:        options.Type,
		Description: options.Description,
		Mode:        types.TransitionMode_Simple,
		Metadata:    map[string]string{},
		Headers:     map[string]string{},
		Properties:  map[string]*types.Value{},
		DataObjects: map[string]*types.Value{},
		Results:     map[string]*types.Value{},
	}

	endpoint.Name = options.Name

	headers, properties, dataObjects, matched := ExtractInOrOut(options.Request)
	if !matched {
		return fmt.Errorf("bad request")
	}
	endpoint.Headers = headers
	endpoint.Properties = properties
	endpoint.DataObjects = dataObjects

	_, properties, _, matched = ExtractInOrOut(options.Response)
	if !matched {
		return fmt.Errorf("bad response")
	}
	endpoint.Results = properties

	name := endpoint.Name
	_, exists := f.endpoints[name]
	if exists {
		return fmt.Errorf("endpoint '%s' already exists", name)
	}
	f.endpoints[name] = endpoint
	return nil
}

func (f *TaskFactoryImpl) ListEndpoint() []types.Endpoint {
	endpoint := make([]types.Endpoint, 0, len(f.endpoints))
	for _, ep := range f.endpoints {
		endpoint = append(endpoint, *ep.DeepCopy())
	}
	return endpoint
}
