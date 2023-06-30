package dns

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

const lookupCheck = "Lookup"
const dialCheck = "Dial"
const timeout = 1 * time.Second
const destination = "ingest.lightstep.com"
const port = "443"

type LookupChecker struct {
}

func (c LookupChecker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	ips, err := net.LookupIP(destination)
	if err != nil {
		return append(results, checks.NewFailedCheck(lookupCheck, "", err))
	} else if len(ips) == 0 {
		return append(results, checks.NewFailedCheck(lookupCheck, "", fmt.Errorf("no ips found")))
	}
	ipStrings := make([]string, len(ips))
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}
	ipString := strings.Join(ipStrings, ",")
	results = append(results, checks.NewSuccessfulCheck(lookupCheck, ipString))

	conn, err := net.DialTimeout("tcp", destination+":"+port, timeout)
	if err != nil {
		return append(results, checks.NewFailedCheck(dialCheck, "failed to connect", err))
	}
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return append(results, checks.NewFailedCheck(dialCheck, "failed to set deadline", err))
	}
	err = conn.Close()
	if err != nil {
		return append(results, checks.NewFailedCheck(dialCheck, "", err))
	}
	return append(results, checks.NewSuccessfulCheck(dialCheck, "can dial ingest.lightstep.com"))
}

func (c LookupChecker) Description() string {
	return "Runs DNS checks to ensure data can be sent"
}

func (c LookupChecker) Name() string {
	return "DNS"
}

func NewLookupCheck(c *checks.Config) checks.Checker {
	return &LookupChecker{}
}
