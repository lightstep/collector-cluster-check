package metrics

import (
	"context"
	"strings"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type ShutdownMeter struct{}

var _ steps.Step = ShutdownMeter{}

func (c ShutdownMeter) Name() string {
	return "ShutdownMeter"
}

func (c ShutdownMeter) Description() string {
	return "Shut down and flush the open meter provider"
}

func (c ShutdownMeter) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	err := deps.MeterProvider.Shutdown(ctx)
	if err != nil && strings.Contains(err.Error(), "DeadlineExceeded") {
		return steps.NewResults(c, steps.NewFailureResultWithHelp(err, "A connection couldn't be established, check firewall rules"))
	} else if err != nil {
		return steps.NewResults(c, steps.NewFailureResultWithHelp(err, "This could mean an incorrect access token was used"))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult("shutdown meter provider"))
}

func (c ShutdownMeter) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.CreateMeterProviderFromConfig(config)}
}
