package dependencies

import (
	"context"
	_ "embed"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	OtelCollectorDependencyName = "OtelCol Config"
)

var (
	//go:embed config.yaml
	collectorConfig string
	podLabels       = map[string]interface{}{
		"app.kubernetes.io/created-by": "collector-cluster-checker",
	}

	// OtelCollectorConfigInitializer generates the Otel Collector CRD configuration
	OtelCollectorConfigInitializer = dependency[*unstructured.Unstructured]{
		dep:     NewOtelConfigClient,
		applier: checks.WithOtelColConfig,
	}
	OtelColMetricInitializer = dependency[*sdkmetric.MeterProvider]{
		dep:     NewOtelColMeterProvider,
		applier: checks.WithMeterProvider,
	}
	OtelColTraceInitializer = dependency[*sdktrace.TracerProvider]{
		dep:     NewOtelColTracerProvider,
		applier: checks.WithTracerProvider,
	}
)

func NewOtelColMeterProvider(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (*sdkmetric.MeterProvider, *checks.Check) {
	return NewMeterProvider(ctx, "localhost:4317", true, false, "", kubeconfig)
}

func NewOtelColTracerProvider(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (*sdktrace.TracerProvider, *checks.Check) {
	return NewTraceProvider(ctx, "localhost:4317", true, false, "", kubeconfig)
}

func NewOtelConfigClient(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (*unstructured.Unstructured, *checks.Check) {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "opentelemetry.io/v1alpha1",
			"kind":       "OpenTelemetryCollector",
			"metadata": map[string]interface{}{
				"name":   "test-col",
				"labels": podLabels,
			},
			"spec": map[string]interface{}{
				"replicas": 1,
				"mode":     "deployment",
				"config":   collectorConfig,
				"env": []map[string]interface{}{
					{
						"name":  "LS_TOKEN",
						"value": token,
					},
					{
						"name":  "DESTINATION",
						"value": endpoint,
					},
				},
			},
		},
	}, checks.NewSuccessfulCheck(OtelCollectorDependencyName, "")
}
