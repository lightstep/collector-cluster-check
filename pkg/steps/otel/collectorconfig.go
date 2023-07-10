package otel

import (
	"context"
	_ "embed"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type CollectorConfig struct {
	token    string
	endpoint string
}

func NewCollectorConfigFromConfig(config *steps.Config) CollectorConfig {
	return CollectorConfig{endpoint: config.Endpoint, token: config.Token}
}

func NewCollectorConfig(token string, endpoint string) *CollectorConfig {
	return &CollectorConfig{token: token, endpoint: endpoint}
}

var _ steps.Step = CollectorConfig{}
var (
	//go:embed config.yaml
	collectorConfig string
	podLabels       = map[string]interface{}{
		"app.kubernetes.io/created-by": "collector-cluster-checker",
	}
)

func (c CollectorConfig) Name() string {
	return "CollectorCRDConfig"
}

func (c CollectorConfig) Description() string {
	return "Gets an initialized collector CRD"
}

func (c CollectorConfig) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	col := &unstructured.Unstructured{
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
						"value": c.token,
					},
					{
						"name":  "DESTINATION",
						"value": c.endpoint,
					},
				},
			},
		},
	}
	return steps.WithOtelColConfig(col), steps.NewSuccessfulResult("retrieved CRD config")
}

func (c CollectorConfig) Dependencies(config *steps.Config) []steps.Step {
	return nil
}
