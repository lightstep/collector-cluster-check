package certmanager

import (
	"context"
	"fmt"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

const (
	issuerCRD               = "issuers.cert-manager.io"
	crdCheck                = "Issuer CRD exists"
	podCheck                = "Cert Manager pod is running"
	certManagerNotInstalled = "cert manager not installed"
	noPodsInstalled         = "no cert manager pods running"
)

type Checker struct {
	client    kubernetes.Interface
	crdClient apiextensionsclientset.Interface
}

func (c Checker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	certManagerCrd, err := c.crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, issuerCRD, metav1.GetOptions{})
	if err != nil {
		return append(results, checks.NewFailedCheck(crdCheck, certManagerNotInstalled, err))
	} else {
		results = append(results, checks.NewSuccessfulCheck(crdCheck, certManagerCrd.Name))
	}
	certManagerPodList, err := c.client.CoreV1().Pods("").List(ctx, v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=cert-manager",
	})
	if err != nil {
		return append(results, checks.NewFailedCheck(podCheck, "", err))
	} else if len(certManagerPodList.Items) == 0 {
		return append(results, checks.NewFailedCheck(podCheck, "", fmt.Errorf(noPodsInstalled)))
	} else {
		podNames := ""
		for _, item := range certManagerPodList.Items {
			podNames = fmt.Sprintf("%s, %s", item.Name, podNames)
		}
		results = append(results, checks.NewSuccessfulCheck(podCheck, podNames))
	}
	return results
}

func (c Checker) Description() string {
	return "Checks that cert manager is installed and running"
}

func (c Checker) Name() string {
	return "Cert Manager"
}

func NewCheck(c *checks.Config) checks.Checker {
	return &Checker{
		client:    c.KubeClient,
		crdClient: c.CustomResourceClient,
	}
}
