package mutate_test

import (
	"github.com/ayoul3/asm-webhook/mutate"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("ContainerHasSecrets", func() {
	Context("When the container is loading the asm secret from regular env", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		container := corev1.Container{
			Image: "nginx",
			Env:   []corev1.EnvVar{{Name: "TEST", Value: "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}},
		}
		//m.K8sClient.CoreV1().ConfigMaps("default").Create(&v1.ConfigMap{Data: map[string]string{}})
		It("should return true", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).To(BeTrue())
		})
	})
})
