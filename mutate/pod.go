package mutate

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *Mutator) MutatePod(ctx context.Context, pod *corev1.Pod) (res *kwhmutating.MutatorResult, err error) {
	var shouldMutate, shouldMutateInit bool
	var containers []corev1.Container

	if containers, shouldMutateInit, err = m.MutateContainers(ctx, pod.Spec.InitContainers, pod); err != nil {
		return res, errors.Wrapf(err, "MutateContainers Initcontainers ")
	}
	log.Debugf("Mutated %d init containers in %s", len(containers), pod.Name)
	pod.Spec.InitContainers = containers

	if containers, shouldMutate, err = m.MutateContainers(ctx, pod.Spec.Containers, pod); err != nil {
		return res, errors.Wrapf(err, "MutateContainers containers ")
	}
	log.Debugf("Mutated mutated %d regular containers in %s", len(containers), pod.Name)
	pod.Spec.Containers = containers

	if shouldMutate || shouldMutateInit {
		pod.Spec.InitContainers = append(m.CreateInitContainer(), pod.Spec.InitContainers...)
		pod.Spec.Volumes = append(pod.Spec.Volumes, m.CreateVolume()...)

	}
	return &kwhmutating.MutatorResult{MutatedObject: pod}, nil
}

func (m *Mutator) CreateInitContainer() []corev1.Container {
	var containers = []corev1.Container{}
	originalPath := fmt.Sprintf("%s%s", m.ASMConfig.OriginalPath, m.ASMConfig.BinaryName)
	newPath := fmt.Sprintf("%s%s", m.ASMConfig.MountPath, m.ASMConfig.BinaryName)
	cmd := fmt.Sprintf("cp %s %s && chmod +x %s", originalPath, newPath, newPath)
	containers = append(containers, corev1.Container{
		Name:            "copy-asm-binary",
		Image:           m.ASMConfig.ImageName,
		ImagePullPolicy: "Always",
		Command:         []string{"sh", "-c", cmd},
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      m.ASMConfig.BinaryName,
				MountPath: m.ASMConfig.MountPath,
			},
		},
	})
	return containers
}

func (m *Mutator) CreateVolume() []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: m.ASMConfig.BinaryName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		},
	}
	return volumes

}

func (m *Mutator) SecretsMutator(ctx context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	log.Debugf(" SecretsMutator - Received object %s in namespace %s", obj.GetName(), obj.GetNamespace())
	switch v := obj.(type) {
	case *corev1.Pod:
		log.Debugf("Got pod %s", v.GetName())
		return m.MutatePod(ctx, v)

	default:
		return &kwhmutating.MutatorResult{}, nil
	}
}
