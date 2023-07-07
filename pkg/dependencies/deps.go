package dependencies

import (
	"context"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

type initFunc[T any] func(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (T, *checks.Check)

type Initializer interface {
	Apply(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (checks.RunnerOption, *checks.Check)
}

type dependency[T any] struct {
	dep     initFunc[T]
	applier func(o T) checks.RunnerOption
}

func (d dependency[T]) Apply(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (checks.RunnerOption, *checks.Check) {
	dep, initResult := d.dep(ctx, endpoint, insecure, http, token, kubeconfig)
	return d.applier(dep), initResult
}
