package observability

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type CloudEventCarrier struct {
	Extension *extensions.DistributedTracingExtension
}

func newCloudEventCarrier() CloudEventCarrier {
	return CloudEventCarrier{Extension: &extensions.DistributedTracingExtension{}}
}

func newCloudEventCarrierWithEvent(event cloudevents.Event) CloudEventCarrier {
	var te, ok = extensions.GetDistributedTracingExtension(event)
	if !ok {
		return CloudEventCarrier{Extension: &extensions.DistributedTracingExtension{}}
	}
	return CloudEventCarrier{Extension: &te}
}

// Get returns the value associated with the passed key.
func (cec CloudEventCarrier) Get(key string) string {
	switch key {
	case extensions.TraceParentExtension:
		return cec.Extension.TraceParent
	case extensions.TraceStateExtension:
		return cec.Extension.TraceState
	default:
		return ""
	}
}

// Set stores the key-value pair.
func (cec CloudEventCarrier) Set(key string, value string) {
	switch key {
	case extensions.TraceParentExtension:
		cec.Extension.TraceParent = value
	case extensions.TraceStateExtension:
		cec.Extension.TraceState = value
	}
}

// Keys lists the keys stored in this carrier.
func (cec CloudEventCarrier) Keys() []string {
	return []string{extensions.TraceParentExtension, extensions.TraceStateExtension}
}

// InjectDistributedTracingExtension injects the tracecontext into the event as a DistributedTracingExtension
func InjectDistributedTracingExtension(ctx context.Context, event cloudevents.Event) {

	// TODO: Should we validate if there's already a tracecontext in the event?
	// Calling it will override any existing value..
	tc := newCloudEventTraceContext()
	carrier := newCloudEventCarrier()
	tc.inject(ctx, carrier)
	carrier.Extension.AddTracingAttributes(&event)
}

// ExtractDistributedTracingExtension reads tracecontext from the cloudevent DistributedTracingExtension into a returned Context.
//
// The returned Context will be a copy of ctx and contain the extracted
// tracecontext as the remote SpanContext. If the extracted tracecontext is
// invalid, the passed ctx will be returned directly instead.
func ExtractDistributedTracingExtension(ctx context.Context, event cloudevents.Event) context.Context {
	tc := newCloudEventTraceContext()
	carrier := newCloudEventCarrierWithEvent(event)

	ctx = tc.extract(ctx, carrier)

	return ctx
}

// cloudEventTraceContext a wrapper around the OpenTelemetry TraceContext
// https://github.com/open-telemetry/opentelemetry-go/blob/main/propagation/trace_context.go
type cloudEventTraceContext struct {
	traceContext propagation.TraceContext
}

func newCloudEventTraceContext() cloudEventTraceContext {
	return cloudEventTraceContext{traceContext: propagation.TraceContext{}}
}

// extract extracts the tracecontext from the cloud event into the context.
//
// If the context has a recording span, then the same context is returned. If not, then the extraction
// from the cloud event happens. The reason for this is to avoid breaking the span order in the trace.
// For instrumented clients, the context *should* have the incoming span from the auto-instrumented library
// thus using this one is more appropriate.
func (etc cloudEventTraceContext) extract(ctx context.Context, carrier CloudEventCarrier) context.Context {
	// TODO: Is there a better way to check if ctx already has a current span on it?
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		// if the context already has an active span so just return that
		return ctx
	}

	// Extracts the traceparent from the cloud event into the context
	// This is useful when there's no context (reading from the queue in a long running process)
	// In this case we can use the traceparent from the event to continue the trace flow.
	return etc.traceContext.Extract(ctx, carrier)
}

// inject injects the current tracecontext from the context into the cloud event
func (etc cloudEventTraceContext) inject(ctx context.Context, carrier CloudEventCarrier) {
	etc.traceContext.Inject(ctx, carrier)
}
