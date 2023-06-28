package checks

import (
	"context"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

type Check struct {
	Name    string
	Message string
	Error   error
}

type CheckerResult []*Check

func (c *Check) IsSuccess() bool {
	return c.Error == nil
}

func (c *Check) IsFailure() bool {
	return c.Error != nil
}

func NewFailedCheck(name string, message string, err error) *Check {
	return &Check{
		Name:    name,
		Message: message,
		Error:   err,
	}
}

func NewSuccessfulCheck(name string, message string) *Check {
	return &Check{
		Name:    name,
		Message: message,
	}
}

type Checker interface {
	Run(ctx context.Context) CheckerResult
	Name() string
	Description() string
}

type Config struct {
	KubeClient           *kubernetes.Clientset
	CustomResourceClient *apiextensionsclientset.Clientset
	MeterProvider        *sdkmetric.MeterProvider
	TracerProvider       *sdktrace.TracerProvider
}

type NewChecker func(c *Config) Checker
type RunnerOption func(r *Runner)

type Runner struct {
	checkers []NewChecker
	conf     *Config
}

func NewRunner(checkers []NewChecker, opts ...RunnerOption) *Runner {
	r := &Runner{
		checkers: checkers,
		conf:     &Config{},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func WithKubernetesClient(client *kubernetes.Clientset) RunnerOption {
	return func(r *Runner) {
		r.conf.KubeClient = client
	}
}

func WithCustomResourceClient(client *apiextensionsclientset.Clientset) RunnerOption {
	return func(r *Runner) {
		r.conf.CustomResourceClient = client
	}
}

func WithMeterProvider(mp *sdkmetric.MeterProvider) RunnerOption {
	return func(r *Runner) {
		r.conf.MeterProvider = mp
	}
}

func WithTracerProvider(tp *sdktrace.TracerProvider) RunnerOption {
	return func(r *Runner) {
		r.conf.TracerProvider = tp
	}
}

func (r *Runner) Run(ctx context.Context) map[string]CheckerResult {
	allResults := map[string]CheckerResult{}
	for _, newChecker := range r.checkers {
		checker := newChecker(r.conf)
		results := checker.Run(ctx)
		allResults[checker.Name()] = append(allResults[checker.Name()], results...)
	}
	return allResults
}
