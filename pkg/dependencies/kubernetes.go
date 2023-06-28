package dependencies

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubernetesDependencyName = "Kubernetes"
)

var (
	KubernetesClientInitializer = dependency[*kubernetes.Clientset]{
		dep:     NewKubernetesClient,
		applier: checks.WithKubernetesClient,
	}
)

func NewKubernetesClient(ctx context.Context, http bool, token string, kubeconfig string) (*kubernetes.Clientset, *checks.Check) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, checks.NewFailedCheck(KubernetesDependencyName, "failed", err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, checks.NewFailedCheck(KubernetesDependencyName, "failed", err)
	}
	return clientset, checks.NewSuccessfulCheck(KubernetesDependencyName, "initialized")
}
