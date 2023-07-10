package otel

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CrdExists struct{}

var _ steps.Step = CrdExists{}

const (
	crdName = "opentelemetrycollectors.opentelemetry.io"
)

func (c CrdExists) Name() string {
	return "CRDExists"
}

func (c CrdExists) Description() string {
	return "checks if the otel collector CRD exists"
}

func (c CrdExists) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	otelCollectorCrd, err := deps.CustomResourceClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.Empty, steps.NewSuccessfulResult(otelCollectorCrd.Name)
}

func (c CrdExists) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{kubernetes.NewCreateCustomResourceClientFromConfig(config)}
}
