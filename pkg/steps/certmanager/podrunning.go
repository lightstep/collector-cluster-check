package certmanager

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodRunning struct{}

var _ steps.Step = PodRunning{}

func (p PodRunning) Name() string {
	return "Cert Manager Running"
}

func (p PodRunning) Description() string {
	return "checks if the cert manager is running"
}

func (p PodRunning) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	operatorPodList, err := deps.KubeClient.CoreV1().Pods("").List(ctx, v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=cert-manager",
	})
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	} else if len(operatorPodList.Items) == 0 {
		return steps.Empty, steps.NewFailureResultWithHelp(nil, "no cert manager pods running")
	}
	podNames := ""
	for _, item := range operatorPodList.Items {
		podNames = fmt.Sprintf("%s, %s", item.Name, podNames)
	}
	return steps.Empty, steps.NewSuccessfulResult(podNames)
}

func (p PodRunning) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{}
}
