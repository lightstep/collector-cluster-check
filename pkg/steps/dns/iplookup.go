package dns

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"net"
)

type IPLookup struct{}

var _ steps.Step = IPLookup{}

const (
	destination = "ingest.lightstep.com"
)

func (c IPLookup) Name() string {
	return "Lightstep IP Lookup"
}

func (c IPLookup) Description() string {
	return "Looks up the IP address of lightstep"
}

func (c IPLookup) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	ips, err := net.LookupIP(destination)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	} else if len(ips) == 0 {
		return steps.Empty, steps.NewFailureResultWithHelp(nil, "no ips found")
	}
	return steps.Empty, steps.NewSuccessfulResult(fmt.Sprintf("%v", ips))
}

func (c IPLookup) Dependencies(config *steps.Config) []steps.Step {
	return nil
}
