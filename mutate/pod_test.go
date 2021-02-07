package mutate_test

import (
	"context"
	"fmt"
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
	RunSpecsWithDefaultAndCustomReporters(t, "asm-webhook - pod", []Reporter{reporters.NewJUnitReporter("pod_report-lib.xml")})
}

func createFakeMutator() mutate.Mutator {
	return mutate.Mutator{
		K8sClient: fake.NewSimpleClientset(),
		Registry: &registry.MockRegistry{
			Image: v1.Config{},
		},
		ASMConfig: mutate.ASMConfig{
			ImageName:    "ayoul3/asm-env",
			MountPath:    "/asm/",
			OriginalPath: "/app/",
			BinaryName:   "asm-env",
		},
	}
}

var _ = Describe("MutatePod", func() {
	m := createFakeMutator()
	Context("When the container has a one command and no args", func() {
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
			execPath := fmt.Sprintf("%s%s", m.ASMConfig.MountPath, m.ASMConfig.BinaryName)
			Expect(v.Spec.Containers[0].Command).To(Equal([]string{execPath}))
			Expect(v.Spec.Containers[0].Args).To(Equal([]string{"/bin/bash"}))
		})
	})
	Context("When the container has a one command and multiple args", func() {
		It("should change the image", func() {
			initialPod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx",
							Env:     []corev1.EnvVar{{Name: "TEST", Value: "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}},
							Command: []string{"/bin/python3"},
							Args:    []string{"script.py", "-c", "arg1"},
						},
					},
				},
			}
			resp, err := m.MutatePod(context.Background(), initialPod)
			Expect(err).ToNot(HaveOccurred())
			v, _ := resp.MutatedObject.(*corev1.Pod)
			execPath := fmt.Sprintf("%s%s", m.ASMConfig.MountPath, m.ASMConfig.BinaryName)
			Expect(v.Spec.Containers[0].Command).To(Equal([]string{execPath}))
			Expect(v.Spec.Containers[0].Args).To(Equal([]string{"/bin/python3", "script.py", "-c", "arg1"}))
		})
	})
	Context("When the container has no command", func() {
		m.Registry.WithImageConfig("test", &v1.Config{Entrypoint: []string{"/bin/sh"}, Cmd: []string{"-c", "echo hello"}})
		It("should change the image", func() {
			initialPod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx",
							Env:     []corev1.EnvVar{{Name: "TEST", Value: "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}},
							Command: nil,
							Args:    nil,
						},
					},
				},
			}
			resp, err := m.MutatePod(context.Background(), initialPod)
			Expect(err).ToNot(HaveOccurred())
			v, _ := resp.MutatedObject.(*corev1.Pod)
			execPath := fmt.Sprintf("%s%s", m.ASMConfig.MountPath, m.ASMConfig.BinaryName)
			Expect(v.Spec.Containers[0].Command).To(Equal([]string{execPath}))
			Expect(v.Spec.Containers[0].Args).To(Equal([]string{"/bin/sh", "-c", "echo hello"}))
		})
	})
	Context("When the container has no command but overwrites image args", func() {
		m.Registry.WithImageConfig("test", &v1.Config{Entrypoint: []string{"/bin/sh"}, Cmd: []string{"-c", "echo hello"}})
		It("should change the image", func() {
			initialPod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx",
							Env:     []corev1.EnvVar{{Name: "TEST", Value: "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}},
							Command: nil,
							Args:    []string{"-c", "echo bonjour"},
						},
					},
				},
			}
			resp, err := m.MutatePod(context.Background(), initialPod)
			Expect(err).ToNot(HaveOccurred())
			v, _ := resp.MutatedObject.(*corev1.Pod)
			execPath := fmt.Sprintf("%s%s", m.ASMConfig.MountPath, m.ASMConfig.BinaryName)
			Expect(v.Spec.Containers[0].Command).To(Equal([]string{execPath}))
			Expect(v.Spec.Containers[0].Args).To(Equal([]string{"/bin/sh", "-c", "echo bonjour"}))
		})
	})
	Context("When no container loads a secret value", func() {
		It("should return the same pod", func() {
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
	Context("When failing to get image config", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
			Registry: &registry.MockRegistry{
				ShouldFail: true,
			},
		}
		It("should return failure", func() {
			initialPod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx", Env: []corev1.EnvVar{{Name: "TEST", Value: "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}}},
					},
				},
			}
			_, err := m.MutatePod(context.Background(), initialPod)
			Expect(err).To(HaveOccurred())
		})
	})
})
