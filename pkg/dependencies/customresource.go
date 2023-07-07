package dependencies

import (
	"context"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

const (
	CRDDependencyName = "CRD"
)

var (
	CustomResourceClientInitializer = dependency[apiextensionsclientset.Interface]{
		dep:     NewCRDClient,
		applier: checks.WithCustomResourceClient,
	}
)

func NewCRDClient(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (apiextensionsclientset.Interface, *checks.Check) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, checks.NewFailedCheck(CRDDependencyName, "failed", err)
	}

	// create the clientset
	clientset, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, checks.NewFailedCheck(CRDDependencyName, "failed", err)
	}
	return clientset, checks.NewSuccessfulCheck(CRDDependencyName, "initialized")
}
