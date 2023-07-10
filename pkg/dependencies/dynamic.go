package dependencies

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	DynamicDependencyName = "Dynamic Client"
)

var (
	DynamicClientInitializer = dependency[dynamic.Interface]{
		dep:     NewDynamicClient,
		applier: checks.WithDynamicClient,
	}
)

func NewDynamicClient(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (dynamic.Interface, *checks.Check) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, checks.NewFailedCheck(DynamicDependencyName, "failed", err)
	}

	// create the clientset
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, checks.NewFailedCheck(DynamicDependencyName, "failed", err)
	}
	return clientset, checks.NewSuccessfulCheck(DynamicDependencyName, "initialized")
}
