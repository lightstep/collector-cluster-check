package dependencies

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
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

var _ steps.Dependency = CreateKubeClient{}

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

func (c CreateKubeClient) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{NewCreateKubeConfigFromConfig(config)}
}

func (c CreateKubeClient) Shutdown(ctx context.Context) error {
	return nil
}
