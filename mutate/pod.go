package mutate

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// MutatePod loops over every initContainer and container to mutate them if necessary
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

// CreateInitContainer injects an initContainer that copies the asm-env binary to a shared mounted volume
func (m *Mutator) CreateInitContainer() []corev1.Container {
	var containers = []corev1.Container{}

	BinPath := fmt.Sprintf("%s%s", m.ASMConfig.BinPath, m.ASMConfig.BinaryName)
	newPath := fmt.Sprintf("%s%s", m.ASMConfig.MountPath, m.ASMConfig.BinaryName)
	cmd := fmt.Sprintf("cp %s %s && chmod +x %s", BinPath, newPath, newPath)
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

// CreateVolume creates the shared volume that receives asm-env binary
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
