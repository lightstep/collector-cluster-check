package dependencies

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"io"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"os"
)

type PortForward struct {
	Port          int
	LabelSelector string
}

var _ steps.Dependency = &PortForward{}

func NewPortForward(port int, labelSelector string) *PortForward {
	return &PortForward{Port: port, LabelSelector: labelSelector}
}

func (p *PortForward) Name() string {
	return fmt.Sprintf("PortForward (%s @ %d)", p.LabelSelector, p.Port)
}

func (p *PortForward) Description() string {
	return "Initiates a port forward"
}

func (p *PortForward) portForwardedResource(conf *steps.Deps, resourceName string) (*steps.PortForwardedResource, error) {
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
	errChan := make(chan error)
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
			errChan <- err
		}
	}()

	select {
	case err = <-errChan:
		close(stopChan)
		close(readyChan)
		return nil, err
	case <-readyChan:
		// If we haven't failed yet, we're okay...
		break
	}

	ports, err := portForwarder.GetPorts()
	if err != nil {
		return nil, err
	}

	return &steps.PortForwardedResource{
		Name:      resourceName,
		LocalPort: int(ports[0].Local),
		Close:     func() { close(stopChan) },
	}, nil
}

func (p *PortForward) Run(ctx context.Context, deps *steps.Deps) (steps.Option, steps.Result) {
	podList, err := deps.DynamicClient.Resource(steps.PodRes).Namespace(apiv1.NamespaceDefault).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		LabelSelector: p.LabelSelector,
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
	return steps.WithPortForwardedResource(pfp), steps.NewSuccessfulResult("started port forward")
}

func (p *PortForward) Shutdown(ctx context.Context) error {
	return nil
}

func (p *PortForward) Dependencies(config *steps.Config) []steps.Dependency {
	return []steps.Dependency{NewCreateKubeClientFromConfig(config)}
}
