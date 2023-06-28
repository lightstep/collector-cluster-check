package certmanager

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	issuerCRD = "issuers.cert-manager.io"
	crdCheck  = "Issuer CRD exists"
	podCheck  = "Cert Manager pod is running"
)

type Checker struct {
	client    *kubernetes.Clientset
	crdClient *apiextensionsclientset.Clientset
}

func (c Checker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	certManagerCrd, err := c.crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, issuerCRD, metav1.GetOptions{})
	if err != nil {
		return append(results, checks.NewFailedCheck(crdCheck, "cert manager not installed", err))
	} else {
		results = append(results, checks.NewSuccessfulCheck(crdCheck, certManagerCrd.Name))
	}
	certManagerPodList, err := c.client.CoreV1().Pods("").List(ctx, v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=cert-manager",
	})
	if err != nil {
		return append(results, checks.NewFailedCheck(podCheck, "", err))
	} else if len(certManagerPodList.Items) == 0 {
		return append(results, checks.NewFailedCheck(podCheck, "", fmt.Errorf("no cert manager pods running")))
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
