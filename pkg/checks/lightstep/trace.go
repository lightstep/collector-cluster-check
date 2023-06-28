package lightstep

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	traceCheck    = "Create Trace"
	endTraceCheck = "Finish Trace"
	traceFlush    = "Flush Traces"
	operationName = "traceChecker.Run"
)

type TraceChecker struct {
	tp *sdktrace.TracerProvider
}

func (c TraceChecker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	_, span := c.tp.Tracer(instrumentation).Start(ctx, operationName)
	results = append(results, checks.NewSuccessfulCheck(traceCheck, ""))
	span.End()
	results = append(results, checks.NewSuccessfulCheck(endTraceCheck, fmt.Sprintf("operation name: %s", operationName)))
	err := c.tp.ForceFlush(ctx)
	if err != nil {
		return append(results, checks.NewFailedCheck(traceFlush, badFlushMessage, err))
	}
	return append(results, checks.NewSuccessfulCheck(traceFlush, "sent Trace"))
}

func (c TraceChecker) Description() string {
	return "Checks that trace data can be sent"
}

func (c TraceChecker) Name() string {
	return "Lightstep Traces"
}

func NewTraceCheck(c *checks.Config) checks.Checker {
	return &TraceChecker{
		tp: c.TracerProvider,
	}
}
