package otel

import (
	"context"
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

func (c QueryCollector) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	r, err := http.Get("http://localhost:8888/metrics")
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.Empty, c.processMetrics(string(data))
}

func (c QueryCollector) processMetrics(metrics string) steps.Result {
	// maps from telemetry type to success ratio
	successMap := map[string]int{}
	failMap := map[string]int{}
	groups := regexp.MustCompile(`otelcol_exporter_sen[td]_(.*)\{.*([0-9]+)`).FindAllStringSubmatch(metrics, -1)
	for _, m := range groups {
		name, count := m[1], m[2]
		i, err := strconv.Atoi(count)
		if err != nil {
			return steps.NewAcceptableFailureResult(err)
		}
		failureGroup := strings.Split(name, "failed_")
		// looking at a failed metric
		if len(failureGroup) == 2 {
			failMap[failureGroup[1]] = i
		} else {
			successMap[name] = i
		}
	}
	// TODO: refactor these
	//for telemetry, count := range successMap {
	//	if failMap[telemetry] > count {
	//		toReturn = append(toReturn, checks.NewFailedCheck(queryMetricsCheck, "", fmt.Errorf("collector failed to send %s", telemetry)))
	//	} else {
	//		toReturn = append(toReturn, checks.NewSuccessfulCheck(queryMetricsCheck, fmt.Sprintf("sent %d %s", count, telemetry)))
	//	}
	//}
	return steps.NewSuccessfulResult("yes")
}

func (c QueryCollector) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{PortForward{Port: 8888}}
}
