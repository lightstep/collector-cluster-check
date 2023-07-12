package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

func TestVersionChecker_Run(t *testing.T) {
	v := Version{}
	tests := []struct {
		name   string
		client kubernetes.Interface
		want   steps.Results
	}{
		{
			name:   "base case",
			client: fake.NewSimpleClientset(),
			want:   steps.NewResults(v, steps.NewSuccessfulResult("v0.0.0-master+$Format:%H$")),
		},
		{
			name: "no client",
			want: steps.NewResults(v, steps.NewFailureResultWithHelp(nil, "client not set")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			deps := &steps.Deps{
				KubeClient: tt.client,
			}
			got := v.Run(ctx, deps)
			assert.Equal(t, tt.want, got)
		})
	}
}
