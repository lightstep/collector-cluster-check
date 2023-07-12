package kubernetes

import (
	"context"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type Version struct{}

var _ steps.Step = Version{}

func (c Version) Name() string {
	return "Kubernetes Version"
}

func (c Version) Description() string {
	return "check kubernetes version"
}

func (c Version) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	if deps.KubeClient == nil {
		return steps.NewResults(c, steps.NewFailureResultWithHelp(nil, "client not set"))
	}
	version, err := deps.KubeClient.Discovery().ServerVersion()
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult(version.String()))
}

func (c Version) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewCreateKubeClientFromConfig(config)}
}
