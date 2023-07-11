package kubernetes

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type StartPortForward struct {
	Port          int
	LabelSelector string
}

var _ steps.Step = StartPortForward{}

func (c StartPortForward) Name() string {
	return "StartPortForward"
}

func (c StartPortForward) Description() string {
	return "Starts a port forward for the label selector at the specified port"
}

func (c StartPortForward) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	return steps.NewResults(c, steps.NewSuccessfulResult(fmt.Sprintf("started port forward @ %d", c.Port)))
}

func (c StartPortForward) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewPortForward(c.Port, c.LabelSelector)}
}
