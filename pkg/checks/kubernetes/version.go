package kubernetes

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"k8s.io/client-go/kubernetes"
)

const versionCheck = "Kubernetes Version"

type VersionChecker struct {
	client *kubernetes.Clientset
}

func (c VersionChecker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	version, err := c.client.ServerVersion()
	if err != nil {
		return append(results, checks.NewFailedCheck(versionCheck, "", err))
	}
	return append(results, checks.NewSuccessfulCheck(versionCheck, version.String()))
}

func (c VersionChecker) Description() string {
	return "Checks the version of the Kubernetes server"
}

func (c VersionChecker) Name() string {
	return "kubernetes version"
}

func NewVersionCheck(c *checks.Config) checks.Checker {
	return &VersionChecker{
		client: c.KubeClient,
	}
}
