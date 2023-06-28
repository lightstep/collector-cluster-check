package dependencies

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

const (
	createMetricExporter = "Metric Exporter"
)

var (
	MetricInitializer = dependency[*sdkmetric.MeterProvider]{
		dep:     NewMeterProvider,
		applier: checks.WithMeterProvider,
	}
)

func NewMeterProvider(ctx context.Context, http bool, token string, kubeconfig string) (*sdkmetric.MeterProvider, *checks.Check) {
	exp, err := newMetricExporter(ctx, http, token)
	if err != nil {
		return nil, checks.NewFailedCheck(createMetricExporter, "", err)
	}
	mp, err := newMetricProvider(exp)
	if err != nil {
		return nil, checks.NewFailedCheck(createMetricExporter, "", err)
	}
	return mp, checks.NewSuccessfulCheck(createMetricExporter, "initialized")
}

func newMetricExporter(ctx context.Context, http bool, token string) (sdkmetric.Exporter, error) {
	var headers = map[string]string{
		"lightstep-access-token": token,
	}
	if http {
		return otlpmetrichttp.New(
			ctx,
			otlpmetrichttp.WithHeaders(headers),
			otlpmetrichttp.WithEndpoint(endpoint),
		)
	} else {
		return otlpmetricgrpc.New(
			ctx,
			otlpmetricgrpc.WithHeaders(headers),
			otlpmetricgrpc.WithEndpoint(endpoint),
		)
	}
}

func newMetricProvider(exp sdkmetric.Exporter) (*sdkmetric.MeterProvider, error) {
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
