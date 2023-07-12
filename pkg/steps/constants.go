package steps

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	ServiceName           = "collector-cluster-check"
	ServiceVersion        = "0.1.0"
	LabelSelector         = "app.kubernetes.io/created-by=collector-cluster-checker"
	CertManagerCrdName    = "issuers.cert-manager.io"
	ServiceMonitorCrdName = "servicemonitors.monitoring.coreos.com"
	OtelCrdName           = "opentelemetrycollectors.opentelemetry.io"
	OtelOperatorSelector  = "app.kubernetes.io/name=opentelemetry-operator"
	CertManagerSelector   = "app.kubernetes.io/name=cert-manager"
)

var (
	ColRes = schema.GroupVersionResource{Group: "opentelemetry.io", Version: "v1alpha1", Resource: "opentelemetrycollectors"}
	PodRes = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
)

type PortForwardedResource struct {
	Name      string
	LocalPort int
	Close     func()
}
