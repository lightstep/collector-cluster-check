package kubernetes

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type PodRunning struct {
	LabelSelector string
}

func NewPodRunning(labelSelector string) *PodRunning {
	return &PodRunning{LabelSelector: labelSelector}
}

var _ steps.Step = PodRunning{}

func (p PodRunning) Name() string {
	return "PodRunning"
}

func (p PodRunning) Description() string {
	return fmt.Sprintf("checks if any pods matching %s are running", p.LabelSelector)
}

func (p PodRunning) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	operatorPodList, err := deps.KubeClient.CoreV1().Pods("").List(ctx, v1.ListOptions{
		LabelSelector: p.LabelSelector,
	})
	if err != nil {
		return steps.NewResults(p, steps.NewFailureResult(err))
	} else if len(operatorPodList.Items) == 0 {
		return steps.NewResults(p, steps.NewFailureResultWithHelp(nil, fmt.Sprintf("no pods matching selector %s running", p.LabelSelector)))
	}
	podNames := ""
	for _, item := range operatorPodList.Items {
		podNames = fmt.Sprintf("%s, %s", item.Name, podNames)
	}
	return steps.NewResults(p, steps.NewSuccessfulResult(podNames))
}

func (p PodRunning) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewCreateKubeClientFromConfig(config)}
}
