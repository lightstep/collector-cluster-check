package kubernetes

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	fakeExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

const (
	fakeCRDName = "test.test.test"
)

func TestCrdExists_Run(t *testing.T) {
	type fields struct {
		CrdName string
	}
	type args struct {
		deps *steps.Deps
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   steps.Results
	}{
		{
			name: "crd found",
			fields: fields{
				CrdName: fakeCRDName,
			},
			args: args{
				deps: &steps.Deps{
					CustomResourceClient: fakeExtensions.NewSimpleClientset(&v1.CustomResourceDefinition{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "apiextensions.k8s.io/v1",
							Kind:       "CustomResourceDefinition",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: fakeCRDName,
						},
					}),
				},
			},
			want: steps.NewResults(CrdExists{}, steps.NewSuccessfulResult(fakeCRDName)),
		},
		{
			name: "crd not found",
			fields: fields{
				CrdName: "test",
			},
			args: args{
				deps: &steps.Deps{
					CustomResourceClient: fakeExtensions.NewSimpleClientset(),
				},
			},
			want: steps.NewResults(CrdExists{}, steps.NewFailureResult(fmt.Errorf("not found"))),
		},
		{
			name: "client not set",
			fields: fields{
				CrdName: "test",
			},
			args: args{
				deps: &steps.Deps{},
			},
			want: steps.NewResults(CrdExists{}, steps.NewFailureResultWithHelp(nil, "custom resource client not set")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CrdExists{
				CrdName: tt.fields.CrdName,
			}
			results := c.Run(context.Background(), tt.args.deps)
			assert.Equal(t, tt.want.StepName(), results.StepName())
			assert.Equal(t, len(tt.want.Steps()), len(results.Steps()))
			for i, result := range results.Steps() {
				if tt.want.Steps()[i].Err() != nil {
					assert.NotNil(t, result.Err())
					assert.ErrorContains(t, result.Err(), result.Err().Error())
				}
				assert.Equal(t, tt.want.Steps()[i].Message(), result.Message())
				assert.Equal(t, tt.want.Steps()[i].Successful(), result.Successful())
				assert.Equal(t, tt.want.Steps()[i].ShouldContinue(), result.ShouldContinue())
			}
		})
	}
}
