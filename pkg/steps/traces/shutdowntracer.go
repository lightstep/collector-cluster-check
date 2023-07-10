package traces

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"strings"
)

type ShutdownTracer struct{}

var _ steps.Step = ShutdownTracer{}

func (c ShutdownTracer) Name() string {
	return "ShutdownTracer"
}

func (c ShutdownTracer) Description() string {
	return "Shut down and flush the open tracer provider"
}

func (c ShutdownTracer) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	err := deps.TracerProvider.Shutdown(ctx)
	if err != nil && strings.Contains(err.Error(), "DeadlineExceeded") {
		return steps.Empty, steps.NewFailureResultWithHelp(err, "A connection couldn't be established, check firewall rules")
	} else if err != nil {
		return steps.Empty, steps.NewFailureResultWithHelp(err, "This could mean an incorrect access token was used")
	}
	return steps.Empty, steps.NewSuccessfulResult("shutdown tracer provider")
}

func (c ShutdownTracer) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{CreateTracerProviderFromConfig(config)}
}
