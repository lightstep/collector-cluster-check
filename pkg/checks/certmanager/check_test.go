package certmanager

import (
	"context"
	"fmt"
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	fakeExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestChecker_Run(t *testing.T) {
	type fields struct {
		client    kubernetes.Interface
		crdClient apiextensionsclientset.Interface
	}
	tests := []struct {
		name   string
		fields fields
		want   checks.CheckerResult
	}{
		{
			name: "Issuer CRD not found",
			fields: fields{
				client:    fake.NewSimpleClientset(),
				crdClient: fakeExtensions.NewSimpleClientset(),
			},
			want: checks.CheckerResult{
				{
					Name:    crdCheck,
					Message: certManagerNotInstalled,
					Error:   fmt.Errorf("no cert manager found"),
				},
			},
		},
		{
			name: "CRD found no operator",
			fields: fields{
				client: fake.NewSimpleClientset(),
				crdClient: fakeExtensions.NewSimpleClientset(&v1.CustomResourceDefinition{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: issuerCRD,
					},
				}),
			},
			want: checks.CheckerResult{
				{
					Name:    crdCheck,
					Message: issuerCRD,
					Error:   nil,
				},
				{
					Name:    podCheck,
					Message: "",
					Error:   fmt.Errorf(noPodsInstalled),
				},
			},
		},
		{
			name: "CRD found pod running",
			fields: fields{
				client: fake.NewSimpleClientset(&corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Pod",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "somewhere",
						Labels: map[string]string{
							"app.kubernetes.io/name": "cert-manager",
						},
					},
					Spec: corev1.PodSpec{},
				}),
				crdClient: fakeExtensions.NewSimpleClientset(&v1.CustomResourceDefinition{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apiextensions.k8s.io/v1",
						Kind:       "CustomResourceDefinition",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: issuerCRD,
					},
				}),
			},
			want: checks.CheckerResult{
				{
					Name:    crdCheck,
					Message: issuerCRD,
					Error:   nil,
				},
				{
					Name:    podCheck,
					Message: "test, ",
					Error:   nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Checker{
				client:    tt.fields.client,
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
