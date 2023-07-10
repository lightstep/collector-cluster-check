package dependencies

import (
	"context"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

var (
	serviceName         = "collector-cluster-check"
	serviceVersion      = "0.1.0"
	createTraceExporter = "Trace exporter"
	TraceInitializer    = dependency[*sdktrace.TracerProvider]{
		dep:     NewTraceProvider,
		applier: checks.WithTracerProvider,
	}
)

func NewTraceProvider(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (*sdktrace.TracerProvider, *checks.Check) {
	exp, err := newTraceExporter(ctx, endpoint, insecure, http, token)
	if err != nil {
		return nil, checks.NewFailedCheck(createTraceExporter, "", err)
	}
	tp, err := newTraceProvider(exp)
	if err != nil {
		return nil, checks.NewFailedCheck(createTraceExporter, "", err)
	}
	return tp, checks.NewSuccessfulCheck(createTraceExporter, "initialized")
}

func newTraceExporter(ctx context.Context, endpoint string, insecure bool, http bool, token string) (*otlptrace.Exporter, error) {
	var headers = map[string]string{
		"lightstep-access-token": token,
	}
	var client otlptrace.Client
	if http {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithEndpoint(endpoint),
		}
		if insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(
			ctx,
			opts...,
		)
	} else {
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithHeaders(headers),
			otlptracegrpc.WithEndpoint(endpoint),
		}
		if insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		return otlptracegrpc.New(
			ctx,
			opts...,
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
