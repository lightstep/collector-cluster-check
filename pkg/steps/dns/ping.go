package dns

import (
	"context"
	"fmt"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

type Ping struct{}

var _ steps.Step = Ping{}

func (c Ping) Name() string {
	return "Ping"
}

func (c Ping) Description() string {
	return "Pings the lightstep address"
}

func (c Ping) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	pinger, err := probing.NewPinger("ingest.lightstep.com")
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	pinger.Count = 3
	pinger.Timeout = timeout
	err = pinger.RunWithContext(ctx)
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	stats := pinger.Statistics()
	if stats.PacketLoss > 0 {
		return steps.NewResults(c, steps.NewFailureResultWithHelp(nil, fmt.Sprintf("%v%% packet loss", stats.PacketLoss)))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult("pong"))
}

func (c Ping) Dependencies(config *steps.Config) []steps.Dependency {
	return nil
}
