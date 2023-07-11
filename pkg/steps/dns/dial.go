package dns

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

type Dial struct{}

var _ steps.Step = Dial{}

const (
	timeout = 1 * time.Second
	port    = "443"
)

func (c Dial) Name() string {
	return "Dial Lightstep"
}

func (c Dial) Description() string {
	return "Dials the lightstep address"
}

func (c Dial) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	conn, err := net.DialTimeout("tcp", destination+":"+port, timeout)
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResultWithHelp(err, "failed to connect"))
	}
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	err = conn.Close()
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	return steps.NewResults(c, steps.NewSuccessfulResult(fmt.Sprintf("can dial %s", destination)))
}

func (c Dial) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{}
}
