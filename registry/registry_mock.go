package registry

import (
	"context"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type MockRegistry struct {
	Image v1.Config
}

func (r *MockRegistry) GetImageConfig(_ context.Context, _ kubernetes.Interface, _ string, _ *corev1.Container, _ *corev1.PodSpec) (*v1.Config, error) {
	return &r.Image, nil
}
