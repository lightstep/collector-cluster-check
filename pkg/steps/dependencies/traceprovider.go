package dependencies

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

type CreateTraceProvider struct {
	endpoint string
	insecure bool
	http     bool
	token    string
}

func CreateTracerProviderFromConfig(config *steps.Config) CreateTraceProvider {
	return CreateTraceProvider{endpoint: config.Endpoint, insecure: config.Insecure, http: config.Http, token: config.Token}
}

func NewCreateTraceProvider(endpoint string, insecure bool, http bool, token string) CreateTraceProvider {
	return CreateTraceProvider{endpoint: endpoint, insecure: insecure, http: http, token: token}
}

var _ steps.Dependency = CreateTraceProvider{}

func (c CreateTraceProvider) Name() string {
	return fmt.Sprintf("Create Trace Provider @ %s", c.endpoint)
}

func (c CreateTraceProvider) Description() string {
	return "Creates a Trace provider"
}

func (c CreateTraceProvider) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	exp, err := c.newTraceExporter(ctx)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	tp, err := c.newTraceProvider(exp)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.WithTracerProvider(tp), steps.NewSuccessfulResult("initialized trace provider")
}

func (c CreateTraceProvider) Dependencies(config *steps.Config) []steps.Dependency {
	return nil
}

func (c CreateTraceProvider) Shutdown(ctx context.Context) error {
	return nil
}

func (c CreateTraceProvider) newTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	var headers = map[string]string{
		"lightstep-access-token": c.token,
	}
	if c.http {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithEndpoint(c.endpoint),
		}
		if c.insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(
			ctx,
			opts...,
		)
	}
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithEndpoint(c.endpoint),
	}
	if c.insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return otlptracegrpc.New(
		ctx,
		opts...,
	)

}

func (c CreateTraceProvider) newTraceProvider(exp *otlptrace.Exporter) (*sdktrace.TracerProvider, error) {
	res, rErr :=
		resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(steps.ServiceName),
				semconv.ServiceVersionKey.String(steps.ServiceVersion),
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
