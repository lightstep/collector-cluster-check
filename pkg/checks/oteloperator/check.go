package oteloperator

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
	crdName          = "opentelemetrycollectors.opentelemetry.io"
	crdCheck         = "Collector CRD exists"
	operatorPodCheck = "otel operator exists"
)

type Checker struct {
	client    *kubernetes.Clientset
	crdClient *apiextensionsclientset.Clientset
}

func (c Checker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check
	otelCollectorCrd, err := c.crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return append(results, checks.NewFailedCheck(crdCheck, "Collector CRD not installed", err))
	} else {
		results = append(results, checks.NewSuccessfulCheck(crdCheck, otelCollectorCrd.Name))
	}
	operatorPodList, err := c.client.CoreV1().Pods("").List(ctx, v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=opentelemetry-operator",
	})
	if err != nil {
		return append(results, checks.NewFailedCheck(operatorPodCheck, "", err))
	} else if len(operatorPodList.Items) == 0 {
		return append(results, checks.NewFailedCheck(operatorPodCheck, "", fmt.Errorf("no otel operator pods running")))
	} else {
		podNames := ""
		for _, item := range operatorPodList.Items {
			podNames = fmt.Sprintf("%s, %s", item.Name, podNames)
		}
		results = append(results, checks.NewSuccessfulCheck(operatorPodCheck, podNames))
	}
	return results
}

func (c Checker) Description() string {
	return "Checks that the otel operator is installed and running"
}

func (c Checker) Name() string {
	return "Otel Operator"
}

func NewCheck(c *checks.Config) checks.Checker {
	return &Checker{
		client:    c.KubeClient,
		crdClient: c.CustomResourceClient,
	}
}