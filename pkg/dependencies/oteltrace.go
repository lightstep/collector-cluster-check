package dependencies

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	serviceName         = "collector-cluster-check"
	serviceVersion      = "0.1.0"
	endpoint            = "ingest.lightstep.com:443"
	createTraceExporter = "Trace exporter"
	TraceInitializer    = dependency[*sdktrace.TracerProvider]{
		dep:     NewTraceProvider,
		applier: checks.WithTracerProvider,
	}
)

func NewTraceProvider(ctx context.Context, http bool, token string, kubeconfig string) (*sdktrace.TracerProvider, *checks.Check) {
	exp, err := newTraceExporter(ctx, http, token)
	if err != nil {
		return nil, checks.NewFailedCheck(createTraceExporter, "", err)
	}
	tp, err := newTraceProvider(exp)
	if err != nil {
		return nil, checks.NewFailedCheck(createTraceExporter, "", err)
	}
	return tp, checks.NewSuccessfulCheck(createTraceExporter, "initialized")
}

func newTraceExporter(ctx context.Context, http bool, token string) (*otlptrace.Exporter, error) {
	var headers = map[string]string{
		"lightstep-access-token": token,
	}
	var client otlptrace.Client
	if http {
		client = otlptracehttp.NewClient(
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithEndpoint(endpoint),
		)
	} else {
		client = otlptracegrpc.NewClient(
			otlptracegrpc.WithHeaders(headers),
			otlptracegrpc.WithEndpoint(endpoint),
		)
	}
	return otlptrace.New(ctx, client)
}

func newTraceProvider(exp *otlptrace.Exporter) (*sdktrace.TracerProvider, error) {
	res, rErr :=
		resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				semconv.ServiceVersionKey.String(serviceVersion),
			),
		)

	if rErr != nil {
		return nil, rErr
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	), nil
}
