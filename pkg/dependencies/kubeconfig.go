package dependencies

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubeConfigDependencyName = "Kube Config Client"
)

var (
	KubeConfigInitializer = dependency[*rest.Config]{
		dep:     NewKubeConfig,
		applier: checks.WithKubeConfig,
	}
)

func NewKubeConfig(ctx context.Context, endpoint string, insecure bool, http bool, token string, kubeconfig string) (*rest.Config, *checks.Check) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, checks.NewFailedCheck(KubeConfigDependencyName, "failed", err)
	}
	return config, checks.NewSuccessfulCheck(KubeConfigDependencyName, "initialized")
}
