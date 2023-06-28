package kubernetes

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

func TestVersionChecker_Run(t *testing.T) {
	tests := []struct {
		name   string
		client kubernetes.Interface
		want   checks.CheckerResult
	}{
		{
			name:   "base case",
			client: fake.NewSimpleClientset(),
			want: []*checks.Check{
				checks.NewSuccessfulCheck(versionCheck, "v0.0.0-master+$Format:%H$"),
			},
		},
		{
			name: "no client",
			want: []*checks.Check{
				checks.NewFailedCheck(versionCheck, "", fmt.Errorf("no client set")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := VersionChecker{
				client: tt.client,
			}
			ctx := context.Background()
			got := c.Run(ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}
