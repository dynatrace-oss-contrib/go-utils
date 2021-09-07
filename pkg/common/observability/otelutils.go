package observability

import (
	"context"
	"log"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// InitOTelTraceProvider configures the OpenTelemetry SDK
//
// The SDK is configured to export the spans to a OpenTelemetry collector via OTLP/GRPC.
func InitOTelTraceProvider(serviceName, collectorGrpcEndpoint string) func() {
	ctx := context.Background()

	_, err := url.ParseRequestURI(collectorGrpcEndpoint)
	if err != nil {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return func() {}
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(), // TODO: This shouldn't be a problem if its expected to be running inside the cluster and not exposed?
		otlptracegrpc.WithEndpoint(collectorGrpcEndpoint),
	)
	if err != nil {
		return func() { log.Printf("Failed to create the trace exporter: %v", err) }
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		if tp == nil {
			return
		}
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down the tracer provider: %v", err)
		}
	}
}
