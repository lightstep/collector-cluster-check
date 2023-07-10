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

type Option func(c *Deps)

type Deps struct {
	KubeClient           kubernetes.Interface
	CustomResourceClient apiextensionsclientset.Interface
	DynamicClient        dynamic.Interface
	MeterProvider        *sdkmetric.MeterProvider
	TracerProvider       *sdktrace.TracerProvider
	OtelColConfig        *unstructured.Unstructured
	KubeConf             *rest.Config
}

func NewDependencies() *Deps {
	return &Deps{}
}

type Config struct {
	Endpoint   string
	Insecure   bool
	Http       bool
	Token      string
	KubeConfig string
}

// Empty is for a step that doesn't change configuration
func Empty(c *Deps) {
	return
}

func WithKubeConfig(conf *rest.Config) Option {
	return func(c *Deps) {
		c.KubeConf = conf
	}
}

func WithDynamicClient(client dynamic.Interface) Option {
	return func(c *Deps) {
		c.DynamicClient = client
	}
}

func WithOtelColConfig(conf *unstructured.Unstructured) Option {
	return func(c *Deps) {
		c.OtelColConfig = conf
	}
}

func WithKubernetesClient(client kubernetes.Interface) Option {
	return func(c *Deps) {
		c.KubeClient = client
	}
}

func WithCustomResourceClient(client apiextensionsclientset.Interface) Option {
	return func(c *Deps) {
		c.CustomResourceClient = client
	}
}

func WithMeterProvider(mp *sdkmetric.MeterProvider) Option {
	return func(c *Deps) {
		c.MeterProvider = mp
	}
}

func WithTracerProvider(tp *sdktrace.TracerProvider) Option {
	return func(c *Deps) {
		c.TracerProvider = tp
	}
}
