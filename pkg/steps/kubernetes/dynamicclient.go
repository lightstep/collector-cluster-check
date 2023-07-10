package kubernetes

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"k8s.io/client-go/dynamic"
)

type CreateDynamicClient struct {
	kubeconfig string
}

func NewCreateDynamicClientFromConfig(config *steps.Config) CreateDynamicClient {
	return CreateDynamicClient{kubeconfig: config.KubeConfig}
}

func NewCreateDynamicClient(kubeconfig string) *CreateDynamicClient {
	return &CreateDynamicClient{kubeconfig: kubeconfig}
}

var _ steps.Step = CreateDynamicClient{}

func (c CreateDynamicClient) Name() string {
	return "CreateDynamicClient"
}

func (c CreateDynamicClient) Description() string {
	return "Creates the dynamic client"
}

func (c CreateDynamicClient) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	// create the clientset
	clientset, err := dynamic.NewForConfig(deps.KubeConf)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.WithDynamicClient(clientset), steps.NewSuccessfulResult("initialize dynamic client")
}

func (c CreateDynamicClient) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{NewCreateKubeConfigFromConfig(config)}
}
