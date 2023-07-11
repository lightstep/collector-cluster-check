package otel

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type QueryCollector struct{}

var _ steps.Step = QueryCollector{}

func (c QueryCollector) Name() string {
	return "QueryCollector"
}

func (c QueryCollector) Description() string {
	return "checks if the otel collector CRD exists"
}

func (c QueryCollector) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	r, err := http.Get("http://localhost:8888/metrics")
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return steps.NewResults(c, steps.NewFailureResult(err))
	}
	return c.processMetrics(string(data))
}

func (c QueryCollector) processMetrics(metrics string) steps.Results {
	// maps from telemetry type to success ratio
	var toReturn []steps.Result
	successMap := map[string]int{}
	failMap := map[string]int{}
	groups := regexp.MustCompile(`otelcol_exporter_sen[td]_(.*)\{.*([0-9]+)`).FindAllStringSubmatch(metrics, -1)
	for _, m := range groups {
		name, count := m[1], m[2]
		i, err := strconv.Atoi(count)
		if err != nil {
			return steps.NewResults(c, steps.NewAcceptableFailureResult(err))
		}
		failureGroup := strings.Split(name, "failed_")
		// looking at a failed metric
		if len(failureGroup) == 2 {
			failMap[failureGroup[1]] = i
		} else {
			successMap[name] = i
		}
	}
	for telemetry, count := range successMap {
		if failMap[telemetry] > count {
			toReturn = append(toReturn, steps.NewFailureResult(fmt.Errorf("collector failed to send %s", telemetry)))
		} else {
			toReturn = append(toReturn, steps.NewSuccessfulResult(fmt.Sprintf("sent %d %s", count, telemetry)))
		}
	}
	if len(toReturn) == 0 {
		toReturn = append(toReturn, steps.NewAcceptableFailureResultWithHelp(nil, "no telemetry metrics found"))
	}
	return steps.NewResults(c, toReturn...)
}

func (c QueryCollector) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{}
}
