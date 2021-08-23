package observability

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/observability"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TODO: What should we put here?
	instrumentationName = "github.com/keptn/go-utils/observability/cloudevents"
)

type OTelObservabilityService struct {
	TracerProvider trace.TracerProvider
	Tracer         trace.Tracer
}

// When creating the cloudevent HTTP Client, by passing the option 'WithObservabilityService', the context
// gets enriched with the incoming parent trace, thus enabling context propagation
func (n OTelObservabilityService) InboundContextDecorators() []func(context.Context, binding.Message) context.Context {
	return []func(context.Context, binding.Message) context.Context{tracePropagatorContextDecorator}
}

// Called by cloudevents internally when an invalid event was received
func (n OTelObservabilityService) RecordReceivedMalformedEvent(ctx context.Context, err error) {

	_, span := n.Tracer.Start(ctx, observability.ClientSpanName, trace.WithAttributes(attribute.String("method", "RecordReceivedMalformedEvent")))
	span.RecordError(err)
	span.End()
}

// Called by cloudevents internally before/after calling the invoker function on the server (StartReceiver)
func (n OTelObservabilityService) RecordCallingInvoker(ctx context.Context, event *cloudevents.Event) (context.Context, func(errOrResult error)) {

	ctx, span := n.Tracer.Start(ctx, observability.ClientSpanName)

	if span.IsRecording() {
		span.SetAttributes(eventSpanAttributes(event, "RecordCallingInvoker")...)
	}

	return ctx, func(errOrResult error) {
		recordSpanError(span, errOrResult)
		span.End()
	}
}

func (n OTelObservabilityService) RecordSendingEvent(ctx context.Context, event cloudevents.Event) (context.Context, func(errOrResult error)) {
	ctx, span := n.Tracer.Start(ctx, observability.ClientSpanName)

	// TODO: Should we add more things here? What about sensitive information?
	if span.IsRecording() {
		span.SetAttributes(eventSpanAttributes(&event, "RecordSendingEvent")...)
	}

	return ctx, func(errOrResult error) {
		recordSpanError(span, errOrResult)
		span.End()
	}
}

func (n OTelObservabilityService) RecordRequestEvent(ctx context.Context, event cloudevents.Event) (context.Context, func(errOrResult error, event *cloudevents.Event)) {

	ctx, span := n.Tracer.Start(ctx, observability.ClientSpanName)

	if span.IsRecording() {
		span.SetAttributes(eventSpanAttributes(&event, "RecordRequestEvent")...)
	}

	return ctx, func(errOrResult error, event *cloudevents.Event) {
		recordSpanError(span, errOrResult)
		span.End()
	}
}

// Returns a OpenTelemetry aware observability service
func NewOTelObservabilityService() *OTelObservabilityService {
	s := &OTelObservabilityService{
		TracerProvider: otel.GetTracerProvider(),
	}

	s.Tracer = s.TracerProvider.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion("1.0.0"), // TODO: Get the package version from somewhere?
	)

	return s
}

// Extracts the traceparent from the msg and creates the proper context to enable propagation
func tracePropagatorContextDecorator(ctx context.Context, msg binding.Message) context.Context {
	var messageCtx context.Context
	if mctx, ok := msg.(binding.MessageContext); ok {
		messageCtx = mctx.Context()
	} else if mctx, ok := binding.UnwrapMessage(msg).(binding.MessageContext); ok {
		messageCtx = mctx.Context()
	}

	if messageCtx == nil {
		return ctx
	}
	span := trace.SpanFromContext(messageCtx)
	if span == nil {
		return ctx
	}
	return trace.ContextWithSpan(ctx, span)
}

func eventSpanAttributes(e *cloudevents.Event, method string) []attribute.KeyValue {
	attr := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String(observability.SpecversionAttr, e.SpecVersion()),
		attribute.String(observability.IdAttr, e.ID()),
		attribute.String(observability.TypeAttr, e.Type()),
		attribute.String(observability.SourceAttr, e.Source()),
	}
	if sub := e.Subject(); sub != "" {
		attr = append(attr, attribute.String(observability.SubjectAttr, sub))
	}
	if dct := e.DataContentType(); dct != "" {
		attr = append(attr, attribute.String(observability.DatacontenttypeAttr, dct))
	}
	return attr
}

func recordSpanError(span trace.Span, errOrResult error) {
	if protocol.IsACK(errOrResult) || !span.IsRecording() {
		return
	}

	var httpResult *cehttp.Result
	if cloudevents.ResultAs(errOrResult, &httpResult) {
		span.RecordError(httpResult)
		if httpResult.StatusCode > 0 {
			span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(httpResult.StatusCode))
		}
	} else {
		span.RecordError(errOrResult)
	}
}
