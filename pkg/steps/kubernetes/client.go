package kubernetes

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"k8s.io/client-go/kubernetes"
)

type CreateKubeClient struct {
	kubeconfig string
}

func NewCreateKubeClientFromConfig(config *steps.Config) CreateKubeClient {
	return CreateKubeClient{kubeconfig: config.KubeConfig}
}

func NewCreateKubeClient(kubeconfig string) *CreateKubeClient {
	return &CreateKubeClient{kubeconfig: kubeconfig}
}

var _ steps.Step = CreateKubeClient{}

func (c CreateKubeClient) Name() string {
	return "CreateKubeClient"
}

func (c CreateKubeClient) Description() string {
	return "Creates the custom resource client"
}

func (c CreateKubeClient) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	// create the clientset
	clientset, err := kubernetes.NewForConfig(deps.KubeConf)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.WithKubernetesClient(clientset), steps.NewSuccessfulResult("initialize dynamic client")
}

func (c CreateKubeClient) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{NewCreateKubeConfigFromConfig(config)}
}
