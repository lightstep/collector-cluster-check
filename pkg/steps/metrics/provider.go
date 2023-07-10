package metrics

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"time"
)

type CreateMeterProvider struct {
	endpoint string
	insecure bool
	http     bool
	token    string
}

func CreateMeterProviderFromConfig(config *steps.Config) CreateMeterProvider {
	return CreateMeterProvider{endpoint: config.Endpoint, insecure: config.Insecure, http: config.Http, token: config.Token}
}

func NewCreateMeterProvider(endpoint string, insecure bool, http bool, token string) CreateMeterProvider {
	return CreateMeterProvider{endpoint: endpoint, insecure: insecure, http: http, token: token}
}

var _ steps.Step = CreateMeterProvider{}

func (c CreateMeterProvider) Name() string {
	return "Create Meter Provider"
}

func (c CreateMeterProvider) Description() string {
	return "Creates a meter provider"
}

func (c CreateMeterProvider) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	exp, err := c.newMetricExporter(ctx)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	mp, err := c.newMetricProvider(exp)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.WithMeterProvider(mp), steps.NewSuccessfulResult("initialized meter provider")
}

func (c CreateMeterProvider) Dependencies(config *steps.Config) []steps.Step {
	return nil
}

func (c CreateMeterProvider) newMetricExporter(ctx context.Context) (sdkmetric.Exporter, error) {
	var headers = map[string]string{
		"lightstep-access-token": c.token,
	}
	if c.http {
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithHeaders(headers),
			otlpmetrichttp.WithEndpoint(c.endpoint),
		}
		if c.insecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		return otlpmetrichttp.New(
			ctx,
			opts...,
		)
	} else {
		opts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithHeaders(headers),
			otlpmetricgrpc.WithEndpoint(c.endpoint),
		}
		if c.insecure {
			opts = append(opts, otlpmetricgrpc.WithInsecure())
		}
		return otlpmetricgrpc.New(
			ctx,
			opts...,
		)
	}
}

func (c CreateMeterProvider) newMetricProvider(exp sdkmetric.Exporter) (*sdkmetric.MeterProvider, error) {
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

	return sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exp,
				sdkmetric.WithInterval(1*time.Second),
				sdkmetric.WithTimeout(5*time.Second),
			),
		),
		sdkmetric.WithResource(res),
	), nil
}
