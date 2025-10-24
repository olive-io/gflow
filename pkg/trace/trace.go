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

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type EmptyExporter struct{}

func (e *EmptyExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (e *EmptyExporter) Shutdown(ctx context.Context) error {
	return nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	//r, err := resource.Merge(
	//	resource.Default(),
	//	resource.NewWithAttributes(
	//		semconv.SchemaURL,
	//		semconv.ServiceName(name),
	//	),
	//)

	//if err != nil {
	//	panic(err)
	//}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
	)
}

func DefaultProvider() *sdktrace.TracerProvider {
	exp := new(EmptyExporter)
	tp := newTraceProvider(exp)
	return tp
}
