package kubernetes

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/lightstep/collector-cluster-check/pkg/steps"
)

const (
	testSelector = "label=thing"
)

func TestPodRunning_Run(t *testing.T) {
	type fields struct {
		LabelSelector string
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
			name: "client not set",
			fields: fields{
				LabelSelector: testSelector,
			},
			args: args{
				deps: &steps.Deps{},
			},
			want: steps.NewResults(PodRunning{}, steps.NewFailureResultWithHelp(nil, "kube client not set")),
		},
		{
			name: "pod not found",
			fields: fields{
				LabelSelector: testSelector,
			},
			args: args{
				deps: &steps.Deps{
					KubeClient: fake.NewSimpleClientset(),
				},
			},
			want: steps.NewResults(PodRunning{}, steps.NewFailureResultWithHelp(nil, fmt.Sprintf("no pods matching selector %s running", testSelector))),
		},
		{
			name: "pod running",
			fields: fields{
				LabelSelector: testSelector,
			},
			args: args{
				deps: &steps.Deps{
					KubeClient: fake.NewSimpleClientset(&corev1.Pod{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Pod",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test",
							Namespace: "somewhere",
							Labels: map[string]string{
								"label": "thing",
							},
						},
						Spec: corev1.PodSpec{},
					}),
				},
			},
			want: steps.NewResults(PodRunning{}, steps.NewSuccessfulResult("test, ")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PodRunning{
				LabelSelector: tt.fields.LabelSelector,
			}
			results := p.Run(context.Background(), tt.args.deps)
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
