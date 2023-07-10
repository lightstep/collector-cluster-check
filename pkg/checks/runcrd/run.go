package runcrd

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"io"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	createCollectorCheck = "Create Collector CRD"
	watchCollectorCheck  = "Watch Collector CRD"
	listPodsCheck        = "List resulting pods"
	portForwardPodCheck  = "Initiate Port Forward"
	queryMetricsCheck    = "Query Collector Metrics"
	deleteCollectorCheck = "Delete Collector CRD"

	metricCheck     = "Create Metric"
	metricFlush     = "Flush Metrics"
	instrumentation = "collector-cluster-check"
	metricName      = "collector.check.alive"

	traceCheck    = "Create Trace"
	endTraceCheck = "Finish Trace"
	traceFlush    = "Flush Traces"
	operationName = "traceChecker.Run"

	badFlushMessage  = "This could mean an incorrect access token was used"
	deadlineExceeded = "A connection couldn't be established, check firewall rules"
)

var (
	colRes = schema.GroupVersionResource{Group: "opentelemetry.io", Version: "v1alpha1", Resource: "opentelemetrycollectors"}
	podRes = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
)

type PortForwardedResource struct {
	Name      string
	LocalPort int
	close     func()
}

type RunCollectorChecker struct {
	client        dynamic.Interface
	collectorConf *unstructured.Unstructured
	tracer        *sdktrace.TracerProvider
	meter         *sdkmetric.MeterProvider
	kubeClient    kubernetes.Interface
	kubeConf      *rest.Config
}

func (c RunCollectorChecker) waitForPodOrTimeout(ctx context.Context, watcher watch.Interface) error {
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

func (c RunCollectorChecker) Run(ctx context.Context) checks.CheckerResult {
	var results []*checks.Check

	// Create col
	fmt.Println("Creating col...")
	res, err := c.client.Resource(colRes).Namespace(apiv1.NamespaceDefault).Create(ctx, c.collectorConf, metav1.CreateOptions{})
	if err != nil {
		return append(results, checks.NewFailedCheck(createCollectorCheck, "", err))
	}
	results = append(results, checks.NewSuccessfulCheck(createCollectorCheck, fmt.Sprintf("%s has been created", res.GetName())))
	w, err := c.kubeClient.CoreV1().Pods(apiv1.NamespaceDefault).Watch(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		LabelSelector: "app.kubernetes.io/created-by=collector-cluster-checker",
	})
	if err != nil {
		return append(results, checks.NewFailedCheck(watchCollectorCheck, "", err))
	}
	err = c.waitForPodOrTimeout(ctx, w)
	if err != nil {
		return append(results, checks.NewFailedCheck(watchCollectorCheck, "", err))
	}
	results = append(results, checks.NewSuccessfulCheck(watchCollectorCheck, "successfully watched pod"))

	podList, err := c.client.Resource(podRes).Namespace(apiv1.NamespaceDefault).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		LabelSelector: "app.kubernetes.io/created-by=collector-cluster-checker",
	})
	if err != nil {
		return append(results, checks.NewFailedCheck(listPodsCheck, "", err))
	} else if len(podList.Items) != 1 {
		return append(results, checks.NewFailedCheck(listPodsCheck, "", fmt.Errorf("no pods found")))
	}
	results = append(results, checks.NewSuccessfulCheck(listPodsCheck, fmt.Sprintf("%s has been retrieved", podList.Items[0].GetName())))

	pfp, err := c.portForwardedResource(4317, 4317, "pods", apiv1.NamespaceDefault, podList.Items[0].GetName())
	if err != nil {
		return append(results, checks.NewFailedCheck(portForwardPodCheck, "", err))
	}
	results = append(results, checks.NewSuccessfulCheck(portForwardPodCheck, fmt.Sprintf("port forwarded on %d", pfp.LocalPort)))

	// Add a sleep to prevent sending too eagerly
	time.Sleep(1 * time.Second)

	counter, err := c.meter.Meter(instrumentation).Int64Counter(metricName)
	if err != nil {
		return append(results, checks.NewFailedCheck(metricCheck, "", err))
	} else {
		results = append(results, checks.NewSuccessfulCheck(metricCheck, fmt.Sprintf("name: %s", metricName)))
	}
	counter.Add(ctx, 1)
	err = c.meter.Shutdown(ctx)
	if err != nil && strings.Contains(err.Error(), "DeadlineExceeded") {
		return append(results, checks.NewFailedCheck(metricFlush, deadlineExceeded, err))
	} else if err != nil {
		return append(results, checks.NewFailedCheck(metricFlush, badFlushMessage, err))
	}
	results = append(results, checks.NewSuccessfulCheck(metricFlush, "sent counter metric"))

	_, span := c.tracer.Tracer(instrumentation).Start(ctx, operationName)
	results = append(results, checks.NewSuccessfulCheck(traceCheck, ""))
	span.End()
	results = append(results, checks.NewSuccessfulCheck(endTraceCheck, fmt.Sprintf("operation name: %s", operationName)))
	err = c.tracer.Shutdown(ctx)
	if err != nil && strings.Contains(err.Error(), "DeadlineExceeded") {
		return append(results, checks.NewFailedCheck(traceFlush, deadlineExceeded, err))
	} else if err != nil {
		return append(results, checks.NewFailedCheck(traceFlush, badFlushMessage, err))
	}
	results = append(results, checks.NewSuccessfulCheck(traceFlush, "sent trace"))
	pfp.close()
	pfp, err = c.portForwardedResource(8888, 8888, "pods", apiv1.NamespaceDefault, podList.Items[0].GetName())
	if err != nil {
		return append(results, checks.NewFailedCheck(portForwardPodCheck, "", err))
	}
	results = append(results, checks.NewSuccessfulCheck(portForwardPodCheck, fmt.Sprintf("port forwarded on %d", pfp.LocalPort)))
	r, err := http.Get("http://localhost:8888/metrics")
	if err != nil {
		return append(results, checks.NewFailedCheck(queryMetricsCheck, "", err))
	}
	results = append(results, checks.NewSuccessfulCheck(queryMetricsCheck, "queried collector metrics"))
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return append(results, checks.NewFailedCheck(queryMetricsCheck, "", err))
	}
	results = append(results, c.processMetrics(string(data))...)

	err = c.client.Resource(colRes).Namespace(apiv1.NamespaceDefault).Delete(ctx, res.GetName(), metav1.DeleteOptions{})
	if err != nil {
		return append(results, checks.NewFailedCheck(deleteCollectorCheck, "", err))
	}
	return append(results, checks.NewSuccessfulCheck(deleteCollectorCheck, fmt.Sprintf("deleted test collector %s", res.GetName())))
}

func (c RunCollectorChecker) processMetrics(metrics string) []*checks.Check {
	var toReturn []*checks.Check
	// maps from telemetry type to success ratio
	successMap := map[string]int{}
	failMap := map[string]int{}
	groups := regexp.MustCompile(`otelcol_exporter_sen[td]_(.*)\{.*([0-9]+)`).FindAllStringSubmatch(metrics, -1)
	for _, m := range groups {
		name, count := m[1], m[2]
		i, err := strconv.Atoi(count)
		if err != nil {
			toReturn = append(toReturn, checks.NewFailedCheck(queryMetricsCheck, "", err))
			continue
		}
		failureGroup := strings.Split(name, "failed_")
		// looking at a failed metric
		if len(failureGroup) == 2 {
			failMap[failureGroup[1]] = i
		} else {
			successMap[name] = i
		}
	}
	for telemetry, count := range successMap {
		if failMap[telemetry] > count {
			toReturn = append(toReturn, checks.NewFailedCheck(queryMetricsCheck, "", fmt.Errorf("collector failed to send %s", telemetry)))
		} else {
			toReturn = append(toReturn, checks.NewSuccessfulCheck(queryMetricsCheck, fmt.Sprintf("sent %d %s", count, telemetry)))
		}
	}
	return toReturn
}

func (c RunCollectorChecker) portForwardedResource(localPort, remotePort int, resource, namespace, resourceName string) (*PortForwardedResource, error) {

	transport, upgrader, err := spdy.RoundTripperFor(c.kubeConf)
	if err != nil {
		return nil, err
	}
	url := c.kubeClient.CoreV1().RESTClient().
		Post().
		Resource(resource).
		Namespace(namespace).
		Name(resourceName).
		SubResource("portforward").
		URL()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	portForwarder, err := portforward.New(
		dialer,
		[]string{fmt.Sprintf("%d:%d", localPort, remotePort)},
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

func (c RunCollectorChecker) Description() string {
	return "Runs DNS checks to ensure data can be sent"
}

func (c RunCollectorChecker) Name() string {
	return "Run Collector"
}

func NewRunCollectorCheck(c *checks.Config) checks.Checker {
	return &RunCollectorChecker{
		client:        c.DynamicClient,
		kubeClient:    c.KubeClient,
		kubeConf:      c.KubeConf,
		collectorConf: c.OtelColConfig,
		tracer:        c.TracerProvider,
		meter:         c.MeterProvider,
	}
}
