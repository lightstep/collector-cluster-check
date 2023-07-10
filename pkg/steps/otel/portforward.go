package otel

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/kubernetes"
	"io"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"log"
	"net/http"
	"os"
)

type PortForward struct {
	Port int
}

type PortForwardedResource struct {
	Name      string
	LocalPort int
	close     func()
}

var _ steps.Step = PortForward{}

func (p PortForward) Name() string {
	return "OpenTelemetry Operator Running"
}

func (p PortForward) Description() string {
	return "checks if the otel operator is running"
}

func (p PortForward) portForwardedResource(conf *steps.Deps, resourceName string) (*PortForwardedResource, error) {

	transport, upgrader, err := spdy.RoundTripperFor(conf.KubeConf)
	if err != nil {
		return nil, err
	}
	url := conf.KubeClient.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Namespace(apiv1.NamespaceDefault).
		Name(resourceName).
		SubResource("portforward").
		URL()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	portForwarder, err := portforward.New(
		dialer,
		[]string{fmt.Sprintf("%d:%d", p.Port, p.Port)},
		stopChan,
		readyChan,
		io.Discard, // Info messages are a little spammy and we don't care.
		os.Stderr,  // Errors we pass through.
	)
	if err != nil {
		return nil, err
	}

	// ForwardPorts is stopped using stopChan.
	go func() {
		if err := portForwarder.ForwardPorts(); err != nil {
			log.Fatalf("ForwardPorts: %v", err)
		}
	}()

	<-readyChan

	ports, err := portForwarder.GetPorts()
	if err != nil {
		return nil, err
	}

	return &PortForwardedResource{
		Name:      resourceName,
		LocalPort: int(ports[0].Local),
		close:     func() { close(stopChan) },
	}, nil
}

func (p PortForward) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	podList, err := deps.DynamicClient.Resource(podRes).Namespace(apiv1.NamespaceDefault).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		LabelSelector: "app.kubernetes.io/created-by=collector-cluster-checker",
	})
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	} else if len(podList.Items) == 0 {
		return steps.Empty, steps.NewFailureResultWithHelp(nil, "no pods found")
	}
	pfp, err := p.portForwardedResource(deps, podList.Items[0].GetName())
	if err != nil {
		return steps.Empty, steps.NewFailureResult(err)
	}
	return steps.Empty, steps.NewSuccessfulResultWithShutdown(fmt.Sprintf("port forwarded %d", p.Port), func(ctx context.Context) error {
		pfp.close()
		return nil
	})
}

func (p PortForward) Dependencies(config *steps.Config) []steps.Step {
	return []steps.Step{kubernetes.NewCreateKubeClientFromConfig(config)}
}
