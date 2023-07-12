package otel

import (
	"context"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type CreateCollector struct{}

var _ steps.Step = CreateCollector{}

func (c CreateCollector) Name() string {
	return "CreateCollector"
}

func (c CreateCollector) Description() string {
	return "checks if the cert manager CRD exists"
}

func (c CreateCollector) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	res, err := deps.DynamicClient.Resource(steps.ColRes).Namespace(apiv1.NamespaceDefault).Create(ctx, deps.OtelColConfig, metav1.CreateOptions{})
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return steps.NewResults(c, steps.NewAcceptableFailureResult(err))
	} else if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult(fmt.Sprintf("%s has been created", res.GetName())))
}

func (c CreateCollector) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewCollectorConfigFromConfig(config), dependencies.NewCreateDynamicClientFromConfig(config)}
}
