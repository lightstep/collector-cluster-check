package lightstep

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	metricCheck     = "Create Metric"
	metricFlush     = "Flush Metrics"
	instrumentation = "collector-cluster-check"
	metricName      = "collector.check.alive"
	badFlushMessage = "This could mean an incorrect access token was used"
)

type MetricChecker struct {
	mp *sdkmetric.MeterProvider
}

func (c MetricChecker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	counter, err := c.mp.Meter(instrumentation).Int64Counter(metricName)
	if err != nil {
		return append(results, checks.NewFailedCheck(metricCheck, "", err))
	} else {
		results = append(results, checks.NewSuccessfulCheck(metricCheck, fmt.Sprintf("name: %s", metricName)))
	}
	counter.Add(ctx, 1)
	err = c.mp.ForceFlush(ctx)
	if err != nil {
		fmt.Println(err)
		return append(results, checks.NewFailedCheck(metricFlush, badFlushMessage, err))
	}
	return append(results, checks.NewSuccessfulCheck(metricFlush, "sent counter metric"))
}

func (c MetricChecker) Description() string {
	return "Checks that metric data can be sent"
}

func (c MetricChecker) Name() string {
	return "Lightstep Metrics"
}

func NewMetricCheck(c *checks.Config) checks.Checker {
	return &MetricChecker{
		mp: c.MeterProvider,
	}
}