package mutate_test

import (
	"context"
	"testing"

	"github.com/ayoul3/asm-webhook/mutate"
	"github.com/ayoul3/asm-webhook/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

func TestLib(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "asm-webhook - Mutate", []Reporter{reporters.NewJUnitReporter("mutate_report-lib.xml")})
}

var _ = Describe("Mutate", func() {
	m := mutate.Mutator{
		K8sClient: fake.NewSimpleClientset(),
		Registry: &registry.MockRegistry{
			Image: v1.Config{},
		},
	}
	Context("When a container has a secret value", func() {
		It("should change the image", func() {
			initialPod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx",
							Env:     []corev1.EnvVar{{Name: "TEST", Value: "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}},
							Command: []string{"/bin/bash"},
							Args:    nil,
						},
					},
				},
			}
			resp, err := m.MutatePod(context.Background(), initialPod)
			Expect(err).ToNot(HaveOccurred())
			v, _ := resp.MutatedObject.(*corev1.Pod)
			Expect(v.Spec.Containers[0].Command).To(Equal([]string{"/asm/asm-env"}))
			Expect(v.Spec.Containers[0].Args).To(Equal([]string{"/bin/bash"}))
		})
	})
	Context("When no container loads a secret value", func() {
		It("should return failure", func() {
			initialPod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx", Env: []corev1.EnvVar{{Name: "TEST", Value: "whatever"}}},
					},
				},
			}
			resp, err := m.MutatePod(context.Background(), initialPod)
			Expect(err).ToNot(HaveOccurred())
			v, _ := resp.MutatedObject.(*corev1.Pod)
			Expect(v).To(Equal(initialPod))
		})
	})
})
