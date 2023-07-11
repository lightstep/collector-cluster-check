package kubernetes

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type FinishPortForward struct {
	Port          int
	LabelSelector string
}

var _ steps.Step = FinishPortForward{}

func (c FinishPortForward) Name() string {
	return "FinishPortForward"
}

func (c FinishPortForward) Description() string {
	return "Completes a port forward for the label selector at the specified port"
}

func (c FinishPortForward) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	deps.PortForward.Close()
	return steps.NewResults(c, steps.NewSuccessfulResult(fmt.Sprintf("Finished port forward @ %d", c.Port)))
}

func (c FinishPortForward) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewPortForward(c.Port, c.LabelSelector)}
}
