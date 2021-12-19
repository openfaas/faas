package tracing

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	jaegerprop "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Exporter string

const (
	JaegerExporter   Exporter = "jaeger"
	LogExporter      Exporter = "log"
	OTELExporter     Exporter = "otlp"
	DisabledExporter Exporter = "disabled"
)

const (
	otelEnvPropagators            = "OTEL_PROPAGATORS"
	otelEnvTraceSExporter         = "OTEL_TRACES_EXPORTER"
	otelEnvExporterLogPrettyPrint = "OTEL_EXPORTER_LOG_PRETTY_PRINT"
	otelEnvExporterLogTimestamps  = "OTEL_EXPORTER_LOG_TIMESTAMPS"
	otelEnvServiceName            = "OTEL_SERVICE_NAME"
	otelExpOTLPProtocol           = "OTEL_EXPORTER_OTLP_PROTOCOL"
)

type Shutdown func(context.Context)

// Provider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func Provider(ctx context.Context, name, version, commit string) (shutdown Shutdown, err error) {
	exporter := Exporter(get(otelEnvTraceSExporter, string(DisabledExporter)))

	var exp tracesdk.TracerProviderOption
	switch exporter {
	case JaegerExporter:
		// configure the collector from the env variables,
		// OTEL_EXPORTER_JAEGER_ENDPOINT/USER/PASSWORD
		// see: https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/jaeger
		j, e := jaeger.New(jaeger.WithCollectorEndpoint())
		exp, err = tracesdk.WithBatcher(j), e
	case LogExporter:
		w := os.Stdout
		opts := []stdouttrace.Option{stdouttrace.WithWriter(w)}
		if truthyEnv(otelEnvExporterLogPrettyPrint) {
			opts = append(opts, stdouttrace.WithPrettyPrint())
		}
		if !truthyEnv(otelEnvExporterLogTimestamps) {
			opts = append(opts, stdouttrace.WithoutTimestamps())
		}

		s, e := stdouttrace.New(opts...)
		exp, err = tracesdk.WithSyncer(s), e
	case OTELExporter:
		// find available env variables for configuration
		// see: https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace#environment-variables
		kind := get(otelExpOTLPProtocol, "grpc")

		var client otlptrace.Client
		switch kind {
		case "grpc":
			client = otlptracegrpc.NewClient()
		case "http":
			client = otlptracehttp.NewClient()
		}
		o, e := otlptrace.New(ctx, client)
		exp, err = tracesdk.WithBatcher(o), e
	default:
		log.Println("tracing disabled")
		// We explicitly DO NOT set the global TracerProvider using otel.SetTracerProvider().
		// The unset TracerProvider returns a "non-recording" span, but still passes through context.
		// return no-op shutdown function
		return func(_ context.Context) {}, nil
	}
	if err != nil {
		return nil, err
	}

	propagators := strings.ToLower(get(otelEnvPropagators, "tracecontext,baggage"))
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(withPropagators(propagators)...),
	)

	resource, err := resource.New(
		context.Background(),
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceVersionKey.String(version),
			attribute.String("service.commit", commit),
			semconv.ServiceNameKey.String(get(otelEnvServiceName, name)),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		exp,
		tracesdk.WithResource(resource),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(provider)

	shutdown = func(ctx context.Context) {
		// Do not let the application hang forever when it is shutdown.
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		err := provider.Shutdown(ctx)
		if err != nil {
			log.Printf("failed to shutdown tracing provider: %v", err)
		}
	}

	return shutdown, nil
}

func truthyEnv(name string) bool {
	value, ok := os.LookupEnv(name)
	if !ok {
		return false
	}

	switch value {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}

func get(name, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return value
}

func withPropagators(propagators string) []propagation.TextMapPropagator {
	out := []propagation.TextMapPropagator{}

	if strings.Contains(propagators, "tracecontext") {
		out = append(out, propagation.TraceContext{})
	}

	if strings.Contains(propagators, "jaeger") {
		out = append(out, jaegerprop.Jaeger{})
	}

	if strings.Contains(propagators, "baggage") {
		out = append(out, propagation.Baggage{})
	}

	return out
}
