package metrics

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

type CreateCounter struct{}

var _ steps.Step = CreateCounter{}

const (
	instrumentation = "collector-cluster-check"
	metricName      = "collector.check.alive"
)

func (c CreateCounter) Name() string {
	return "CreateCounter"
}

func (c CreateCounter) Description() string {
	return "creates an otel counter"
}

func (c CreateCounter) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	counter, err := deps.MeterProvider.Meter(instrumentation).Int64Counter(metricName)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	counter.Add(ctx, 1)
	return steps.Empty, steps.NewSuccessfulResult(fmt.Sprintf("incremented counter %s", metricName))
}

func (c CreateCounter) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{CreateMeterProviderFromConfig(config)}
}
