package tracing

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Middleware returns a http.HandlerFunc that initializes and replaces the OpenTelemetry span for each request.
func Middleware(nameFormatter func(r *http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	_, ok := os.LookupEnv("OTEL_EXPORTER")
	if !ok {
		return next
	}
	log.Println("configuring proxy tracing middleware")

	propagator := otel.GetTextMapPropagator()

	return func(w http.ResponseWriter, r *http.Request) {
		// get the parent span from the request headers
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		opts := []trace.SpanStartOption{
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
			trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("gateway", "", r)...),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		ctx, span := otel.Tracer("Gateway").Start(ctx, nameFormatter(r), opts...)
		defer span.End()

		debug(span, "tracing request %q", r.URL.String())

		r = r.WithContext(ctx)
		// set the new span as the parent span in the outgoing request context
		// note that this will overwrite the uber-trace-id and traceparent headers
		propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))
		next(w, r)
	}
}

// ConstantName geneates the given name for the span based on the request.
func ConstantName(value string) func(*http.Request) string {
	return func(r *http.Request) string {
		return value
	}
}

func debug(span trace.Span, format string, args ...interface{}) {
	value := os.Getenv("OTEL_LOG_LEVEL")
	if strings.ToLower(value) != "debug" {
		return
	}

	log.Printf("%s, trace_id=%s", fmt.Sprintf(format, args...), span.SpanContext().TraceID())
}
