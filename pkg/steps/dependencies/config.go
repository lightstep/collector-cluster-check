package dependencies

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"k8s.io/client-go/tools/clientcmd"
)

type CreateKubeConfig struct {
	kubeconfig string
}

func NewCreateKubeConfigFromConfig(config *steps.Config) CreateKubeConfig {
	return CreateKubeConfig{kubeconfig: config.KubeConfig}
}

func NewCreateKubeConfig(kubeconfig string) CreateKubeConfig {
	return CreateKubeConfig{kubeconfig: kubeconfig}
}

var _ steps.Dependency = CreateKubeConfig{}

func (c CreateKubeConfig) Name() string {
	return "CreateKubeConfig"
}

func (c CreateKubeConfig) Description() string {
	return "Creates the kube config"
}

func (c CreateKubeConfig) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	// use the current context in KubeConfig
	config, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.WithKubeConfig(config), steps.NewSuccessfulResult("initialize Kube Config")
}

func (c CreateKubeConfig) Dependencies(config *steps.Config) []steps.Dependency {
	return nil
}

func (c CreateKubeConfig) Shutdown(ctx context.Context) error {
	return nil
}
