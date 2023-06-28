package prometheus

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	fakeExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
)

func TestChecker_Run(t *testing.T) {
	type fields struct {
		crdClient apiextensionsclientset.Interface
	}
	tests := []struct {
		name   string
		fields fields
		want   checks.CheckerResult
	}{
		{
			name: "Service Monitor CRD not found",
			fields: fields{
				crdClient: fakeExtensions.NewSimpleClientset(),
			},
			want: checks.CheckerResult{
				{
					Name:    crdCheck,
					Message: serviceMonitorNotInstalled,
					Error:   fmt.Errorf("no Service Monitor CRD found"),
				},
			},
		},
		{
			name: "CRD found",
			fields: fields{
				crdClient: fakeExtensions.NewSimpleClientset(&v1.CustomResourceDefinition{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: serviceMonitorCRDName,
					},
				}),
			},
			want: checks.CheckerResult{
				{
					Name:    crdCheck,
					Message: serviceMonitorCRDName,
					Error:   nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Checker{
				crdClient: tt.fields.crdClient,
			}
			ctx := context.Background()
			got := c.Run(ctx)
			assert.Equal(t, len(got), len(tt.want))
			for i, check := range got {
				assert.Equal(t, check.Name, tt.want[i].Name)
				assert.Equal(t, check.Message, tt.want[i].Message)
				assert.Equal(t, check.IsFailure(), tt.want[i].IsFailure())
			}
		})
	}
}
