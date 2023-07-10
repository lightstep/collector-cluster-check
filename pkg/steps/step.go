package steps

import (
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Result struct {
	// successful if the check completed without error
	successful bool

	// err is any error that the step encountered
	err error

	// shouldContinue is whether the check should continue
	shouldContinue bool
}

type Config struct {
	KubeClient           kubernetes.Interface
	CustomResourceClient apiextensionsclientset.Interface
	DynamicClient        dynamic.Interface
	MeterProvider        *sdkmetric.MeterProvider
	TracerProvider       *sdktrace.TracerProvider
	OtelColConfig        *unstructured.Unstructured
	KubeConf             *rest.Config
}

type Option func(c *Config)
type NewStep func(c *Config) Step

type Step interface {
	// Name is a single word identifier
	Name() string

	// Description is the optional explanation for what the step does
	Description() string

	Run() Result

	// Dependencies is a list of steps that must be run prior to this step
	Dependencies() []Step
}
