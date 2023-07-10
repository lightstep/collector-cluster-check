package otel

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/kubernetes"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CreateCollector struct{}

var _ steps.Step = CreateCollector{}

var (
	colRes = schema.GroupVersionResource{Group: "opentelemetry.io", Version: "v1alpha1", Resource: "opentelemetrycollectors"}
	podRes = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
)

func (c CreateCollector) Name() string {
	return "CreateCollector"
}

func (c CreateCollector) Description() string {
	return "checks if the cert manager CRD exists"
}

func (c CreateCollector) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	res, err := deps.DynamicClient.Resource(colRes).Namespace(apiv1.NamespaceDefault).Create(ctx, deps.OtelColConfig, metav1.CreateOptions{})
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.Empty, steps.NewSuccessfulResult(fmt.Sprintf("%s has been created", res.GetName()))
}

func (c CreateCollector) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{NewCollectorConfigFromConfig(config), kubernetes.NewCreateDynamicClientFromConfig(config)}
}
