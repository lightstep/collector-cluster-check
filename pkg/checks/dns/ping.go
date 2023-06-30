package dns

import (
	"context"
	"fmt"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

const pingCheck = "Ping"

type PingChecker struct {
}

func (c PingChecker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	pinger, err := probing.NewPinger("ingest.lightstep.com")
	if err != nil {
		return append(results, checks.NewFailedCheck(pingCheck, "", err))
	}
	pinger.Count = 3
	pinger.Timeout = timeout
	err = pinger.RunWithContext(ctx)
	if err != nil {
		return append(results, checks.NewFailedCheck(pingCheck, "", err))
	}
	stats := pinger.Statistics()
	if stats.PacketLoss > 0 {
		return append(results, checks.NewFailedCheck(pingCheck, fmt.Sprintf("%v%% packet loss", stats.PacketLoss), fmt.Errorf("too many packets lost")))
	}
	return append(results, checks.NewSuccessfulCheck(pingCheck, "pong"))
}

func (c PingChecker) Description() string {
	return "Runs DNS checks to ensure data can be sent"
}

func (c PingChecker) Name() string {
	return "DNS"
}

func NewDialCheck(c *checks.Config) checks.Checker {
	return &PingChecker{}
}
