package registry

import (
	"context"
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type MockRegistry struct {
	Image      v1.Config
	ShouldFail bool
}

func (r *MockRegistry) GetImageConfig(_ context.Context, _ kubernetes.Interface, _ string, _ *corev1.Container, _ *corev1.PodSpec) (*v1.Config, error) {
	if r.ShouldFail {
		return nil, fmt.Errorf("error GetImageConfig")
	}
	return &r.Image, nil
}

func (r *MockRegistry) WithImageConfig(name string, imageConfig *v1.Config) {
	r.Image = *imageConfig
}
