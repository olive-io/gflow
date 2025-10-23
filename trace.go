package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

type MyExporter struct{}

func (m *MyExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (m *MyExporter) Shutdown(ctx context.Context) error {
	return nil
}

func newExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	return new(MyExporter), nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// 确保默认的 SDK 资源和所需的服务名称已设置。
	//r, err := resource.Merge(
	//	resource.Default(),
	//	resource.NewWithAttributes(
	//		semconv.SchemaURL,
	//		semconv.ServiceName("ExampleService"),
	//	),
	//)

	//if err != nil {
	//	panic(err)
	//}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		//sdktrace.WithResource(r),
	)
}

func main() {
	ctx := context.Background()

	exp, err := newExporter(ctx)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	// 创建一个跟踪器提供程序，并使用给定的 exporter、批量 span 处理器。
	tp := newTraceProvider(exp)

	// 正确处理关闭操作以避免资源泄漏。
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)

	// 最后，设置可用于该包的跟踪器。
	tracer = tp.Tracer("ExampleService")

	ctx, span := tracer.Start(ctx, "start process")
	defer span.End()

	toChildTask(ctx)
}

func toChildTask(ctx context.Context) {
	sc := trace.SpanContextConfig{
		TraceID:    trace.TraceID{},
		SpanID:     trace.SpanID{},
		TraceFlags: 0,
		TraceState: trace.TraceState{},
		Remote:     false,
	}
	spanCtx := trace.NewSpanContext(sc)
	spanCtx = trace.SpanContextFromContext(ctx)
	ctx = trace.ContextWithSpanContext(context.TODO(), spanCtx)

	//span := trace.SpanFromContext(ctx)
	//defer span.End()
	ctx, span := tracer.Start(ctx, "task 1")
	defer span.End()
	tracer.Start(ctx, "task 2", trace.WithSpanKind(trace.SpanKindClient))

	span.AddEvent("do task")
}
