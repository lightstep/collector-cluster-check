package traces

import (
	"context"
	"strings"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type ShutdownTracer struct{}

var _ steps.Step = ShutdownTracer{}

func (c ShutdownTracer) Name() string {
	return "ShutdownTracer"
}

func (c ShutdownTracer) Description() string {
	return "Shut down and flush the open tracer provider"
}

func (c ShutdownTracer) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	err := deps.TracerProvider.Shutdown(ctx)
	if err != nil && strings.Contains(err.Error(), "DeadlineExceeded") {
		return steps.NewResults(c, steps.NewFailureResultWithHelp(err, "A connection couldn't be established, check firewall rules"))
	} else if err != nil {
		return steps.NewResults(c, steps.NewFailureResultWithHelp(err, "This could mean an incorrect access token was used"))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult("shutdown tracer provider"))
}

func (c ShutdownTracer) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.CreateTracerProviderFromConfig(config)}
}
