package metrics

import (
	"context"
	"fmt"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type CreateCounter struct {
	endpoint string
	insecure bool
}

func NewCreateCounter(endpoint string, insecure bool) CreateCounter {
	return CreateCounter{endpoint: endpoint, insecure: insecure}
}

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

func (c CreateCounter) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	counter, err := deps.MeterProvider.Meter(instrumentation).Int64Counter(metricName)
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	counter.Add(ctx, 1)
	return steps.NewResults(c, steps.NewSuccessfulResult(fmt.Sprintf("incremented counter %s", metricName)))
}

func (c CreateCounter) Dependencies(config *steps.Config) []steps.Dependency {
	if len(c.endpoint) > 0 {
		return []steps.Dependency{dependencies.NewCreateMeterProvider(c.endpoint, c.insecure, config.Http, config.Token)}
	}
	return []steps.Dependency{dependencies.CreateMeterProviderFromConfig(config)}
}
