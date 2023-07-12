package otel

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type DeleteCollector struct{}

var _ steps.Step = DeleteCollector{}

func (c DeleteCollector) Name() string {
	return "DeleteCollector"
}

func (c DeleteCollector) Description() string {
	return "checks if the cert manager CRD exists"
}

func (c DeleteCollector) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	err := deps.DynamicClient.Resource(steps.ColRes).Namespace(apiv1.NamespaceDefault).Delete(ctx, deps.OtelColConfig.GetName(), metav1.DeleteOptions{})
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult(fmt.Sprintf("%s has been deleted", deps.OtelColConfig.GetName())))
}

func (c DeleteCollector) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewCollectorConfigFromConfig(config), dependencies.NewCreateDynamicClientFromConfig(config)}
}
