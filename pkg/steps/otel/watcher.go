package otel

import (
	"context"
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dependencies"
)

type PodWatcher struct{}

var _ steps.Step = PodWatcher{}

func (p PodWatcher) Name() string {
	return "PodWatcher"
}

func (p PodWatcher) Description() string {
	return "checks if the collector pod is running"
}

func (p PodWatcher) waitForPodOrTimeout(ctx context.Context, watcher watch.Interface) error {
	ctxTimeout, cancelFunc := context.WithTimeout(ctx, time.Second*5)
	defer cancelFunc()
	defer watcher.Stop()
	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				return fmt.Errorf("error watching")
			}
			p, ok := event.Object.(*apiv1.Pod)
			if !ok {
				return fmt.Errorf("unexpected type")
			}
			if p.Status.Phase == apiv1.PodRunning {
				return nil
			}
		case <-ctxTimeout.Done():
			return fmt.Errorf("timeout while waiting")
		}
	}
}

func (p PodWatcher) Run(ctx context.Context, deps *steps.Deps) steps.Results {
	w, err := deps.KubeClient.CoreV1().Pods(apiv1.NamespaceDefault).Watch(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		LabelSelector: "app.kubernetes.io/created-by=collector-cluster-checker",
	})
	if err != nil {
		return steps.NewResults(p, steps.NewFailureResult(err))
	}
	err = p.waitForPodOrTimeout(ctx, w)
	if err != nil {
		return steps.NewResults(p, steps.NewFailureResult(err))
	}
	return steps.NewResults(p, steps.NewSuccessfulResult("successfully waited for running pod"))
}

func (p PodWatcher) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{dependencies.NewCreateKubeClientFromConfig(config)}
}
