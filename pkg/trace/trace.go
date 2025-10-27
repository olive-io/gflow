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

package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	Endpoint string        `json:"endpoint" yaml:"endpoint"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
	Insecure bool          `json:"insecure" yaml:"insecure"`
}

func NewJaegerTraceProvider(ctx context.Context, cfg *Config) (*sdktrace.TracerProvider, error) {
	var opts []otlptracehttp.Option

	if cfg.Endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	traceExporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	var providerOptions []sdktrace.TracerProviderOption
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	batcher := sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(timeout))
	providerOptions = append(providerOptions, batcher)

	traceProvider := sdktrace.NewTracerProvider(providerOptions...)
	return traceProvider, nil
}

type EmptyExporter struct{}

func (e *EmptyExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (e *EmptyExporter) Shutdown(ctx context.Context) error {
	return nil
}

func DefaultProvider() *sdktrace.TracerProvider {
	exp := new(EmptyExporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)
	return tp
}
