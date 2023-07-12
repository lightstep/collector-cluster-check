package traces

import (
	"context"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type StartTrace struct {
	endpoint string
	insecure bool
}

func NewStartTrace(endpoint string, insecure bool) StartTrace {
	return StartTrace{endpoint: endpoint, insecure: insecure}
}

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

func (c StartTrace) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	_, span := deps.TracerProvider.Tracer(instrumentation).Start(ctx, operationName)
	span.End()
	return steps.NewResults(c, steps.NewSuccessfulResult("started and ended trace"))
}

func (c StartTrace) Dependencies(config *steps.Config) []steps.Dependency {
	if len(c.endpoint) > 0 {
		return []steps.Dependency{dependencies.NewCreateTraceProvider(c.endpoint, c.insecure, config.Http, config.Token)}
	}
	return []steps.Dependency{dependencies.CreateTracerProviderFromConfig(config)}
}
