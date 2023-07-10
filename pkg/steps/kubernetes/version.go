package kubernetes

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

type Version struct{}

var _ steps.Step = Version{}

func (c Version) Name() string {
	return "Kubernetes Version"
}

func (c Version) Description() string {
	return "check kubernetes version"
}

func (c Version) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	version, err := deps.KubeClient.Discovery().ServerVersion()
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.Empty, steps.NewSuccessfulResult(version.String())
}

func (c Version) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{NewCreateKubeClientFromConfig(config)}
}
