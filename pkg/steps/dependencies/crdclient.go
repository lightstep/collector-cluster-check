package dependencies

import (
	"context"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

type CreateCustomResourceClient struct {
	kubeconfig string
}

func NewCreateCustomResourceClientFromConfig(config *steps.Config) CreateCustomResourceClient {
	return CreateCustomResourceClient{kubeconfig: config.KubeConfig}
}

func NewCreateCustomResourceClient(kubeconfig string) *CreateCustomResourceClient {
	return &CreateCustomResourceClient{kubeconfig: kubeconfig}
}

var _ steps.Dependency = CreateCustomResourceClient{}

func (c CreateCustomResourceClient) Name() string {
	return "CreateCustomResourceClient"
}

func (c CreateCustomResourceClient) Description() string {
	return "Creates the custom resource client"
}

func (c CreateCustomResourceClient) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	// create the clientset
	clientset, err := apiextensionsclientset.NewForConfig(deps.KubeConf)
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.WithCustomResourceClient(clientset), steps.NewSuccessfulResult("initialize CRD client")
}

func (c CreateCustomResourceClient) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{NewCreateKubeConfigFromConfig(config)}
}

func (c CreateCustomResourceClient) Shutdown(ctx context.Context) error {
	return nil
}
