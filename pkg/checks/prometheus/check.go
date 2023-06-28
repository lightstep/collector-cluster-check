package prometheus

import (
	"context"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	serviceMonitorCRDName = "servicemonitors.monitoring.coreos.com"
	crdCheck              = "Service Monitor CRD exists"
)

type Checker struct {
	crdClient *apiextensionsclientset.Clientset
}

func (c Checker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	serviceMonitorCRD, err := c.crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, serviceMonitorCRDName, metav1.GetOptions{})
	if err != nil {
		return append(results, checks.NewFailedCheck(crdCheck, "Service Monitor CRD not installed", err))
	}
	return append(results, checks.NewSuccessfulCheck(crdCheck, serviceMonitorCRD.Name))
}

func (c Checker) Description() string {
	return "Checks that the service monitor CRD is installed"
}

func (c Checker) Name() string {
	return "Prometheus"
}

func NewCheck(c *checks.Config) checks.Checker {
	return &Checker{
		crdClient: c.CustomResourceClient,
	}
}