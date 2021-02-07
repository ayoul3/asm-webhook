package mutate_test

import (
	"context"

	"github.com/ayoul3/asm-webhook/mutate"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("ContainerHasSecrets", func() {
	Context("When the asm secret is in regular env", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		container := corev1.Container{
			Image: "nginx",
			Env:   []corev1.EnvVar{{Name: "TEST", Value: "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}},
		}
		It("should return true", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).To(BeTrue())
		})
	})
	Context("When the asm secret is sourced in env configmap", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		configMap, _ := m.K8sClient.CoreV1().ConfigMaps("default").Create(
			context.Background(), &corev1.ConfigMap{Data: map[string]string{"key": "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}}, v1.CreateOptions{},
		)
		container := corev1.Container{
			Image: "nginx",
			Env: []corev1.EnvVar{{Name: "TEST", ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: configMap.Name}},
			}}},
		}

		It("should return true", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).To(BeTrue())
		})
	})
	Context("When the asm secret is sourced in env Secret ref", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		secretRef, _ := m.K8sClient.CoreV1().Secrets("default").Create(
			context.Background(),
			&corev1.Secret{Data: map[string][]byte{"key": []byte(`arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP`)}},
			v1.CreateOptions{},
		)
		container := corev1.Container{
			Image: "nginx",
			Env: []corev1.EnvVar{{Name: "TEST", ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: secretRef.Name}},
			}}},
		}

		It("should return true", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).To(BeTrue())
		})
	})
	Context("When the configmap is not found", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		configMap, _ := m.K8sClient.CoreV1().ConfigMaps("different").Create(
			context.Background(), &corev1.ConfigMap{Data: map[string]string{"key": "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}}, v1.CreateOptions{},
		)
		container := corev1.Container{
			Image: "nginx",
			Env: []corev1.EnvVar{{Name: "TEST", ValueFrom: &corev1.EnvVarSource{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: configMap.Name}},
			}}},
		}

		It("should not return an error", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).ToNot(BeTrue())
		})
	})
	Context("When the asm secret is in envFrom configmap", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		configMap, _ := m.K8sClient.CoreV1().ConfigMaps("default").Create(
			context.Background(), &corev1.ConfigMap{Data: map[string]string{"key": "arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP"}}, v1.CreateOptions{},
		)
		container := corev1.Container{
			Image: "nginx",
			EnvFrom: []corev1.EnvFromSource{
				{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: configMap.Name}}},
			},
		}

		It("should return true", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).To(BeTrue())
		})
	})
	Context("When the asm secret is in envFrom secretref", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		secretRef, _ := m.K8sClient.CoreV1().Secrets("default").Create(
			context.Background(),
			&corev1.Secret{Data: map[string][]byte{"key": []byte(`arn:aws:secretsmanager:eu-west-1:886477354405:secret:/key1-mIdVIP`)}},
			v1.CreateOptions{},
		)
		container := corev1.Container{
			Image: "nginx",
			EnvFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: secretRef.Name}}},
			},
		}

		It("should return true", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).To(BeTrue())
		})
	})
	Context("When there is no secret anywhere", func() {
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		configMap, _ := m.K8sClient.CoreV1().ConfigMaps("default").Create(
			context.Background(), &corev1.ConfigMap{Data: map[string]string{"key2": "value2"}}, v1.CreateOptions{},
		)
		secretRef, _ := m.K8sClient.CoreV1().Secrets("default").Create(
			context.Background(),
			&corev1.Secret{Data: map[string][]byte{"key1": []byte(`value1`)}},
			v1.CreateOptions{},
		)
		container := corev1.Container{
			Image: "nginx",
			EnvFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: secretRef.Name}}},
				{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: configMap.Name}}},
			},
			Env: []corev1.EnvVar{
				{
					Name:      "TEST1",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: secretRef.Name}}},
				},
				{
					Name:      "TEST2",
					ValueFrom: &corev1.EnvVarSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: configMap.Name}}},
				},
			},
		}

		It("should return false", func() {
			hasSecrets, err := m.ContainerHasSecrets(&container, "default")
			Expect(err).ToNot(HaveOccurred())
			Expect(hasSecrets).ToNot(BeTrue())
		})
	})
})
