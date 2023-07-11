package kubernetes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type CrdExists struct {
	CrdName string
}

func NewCrdExists(crdName string) CrdExists {
	return CrdExists{CrdName: crdName}
}

var _ steps.Step = CrdExists{}

func (c CrdExists) Name() string {
	return "CRDExists"
}

func (c CrdExists) Description() string {
	return "checks if the specified CRD exists"
}

func (c CrdExists) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	retrievedCrd, err := deps.CustomResourceClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, c.CrdName, metav1.GetOptions{})
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult(retrievedCrd.Name))
}

func (c CrdExists) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewCreateCustomResourceClientFromConfig(config)}
}
