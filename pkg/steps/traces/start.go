package traces

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

type StartTrace struct{}

var _ steps.Step = StartTrace{}

const (
	instrumentation = "collector-cluster-check"
	operationName   = "traceChecker.Run"
)

func (c StartTrace) Name() string {
	return "Start trace"
}

func (c StartTrace) Description() string {
	return "Starts an otel trace"
}

func (c StartTrace) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	_, span := deps.TracerProvider.Tracer(instrumentation).Start(ctx, operationName)
	defer span.End()
	return steps.Empty, steps.NewSuccessfulResult("started and ended trace")
}

func (c StartTrace) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{CreateTracerProviderFromConfig(config)}
}
