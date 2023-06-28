package dependencies

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"time"
)

const (
	CreateMetricExporter = "Metric Exporter"
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
		return nil, checks.NewFailedCheck(CreateMetricExporter, "", err)
	}
	return newMetricProvider(exp), checks.NewSuccessfulCheck(CreateMetricExporter, "initialized")
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

func newMetricProvider(exp sdkmetric.Exporter) *sdkmetric.MeterProvider {
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
		panic(rErr)
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
	)
}
