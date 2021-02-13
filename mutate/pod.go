package mutate

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// MutatePod loops over every initContainer and container to mutate them if necessary
func (m *Mutator) MutatePod(ctx context.Context, pod *corev1.Pod, config ASMConfig) (res *kwhmutating.MutatorResult, err error) {
	var shouldMutate, shouldMutateInit bool
	var containers []corev1.Container

	config.Log.Debugf("Looking at %d initContainers", len(pod.Spec.InitContainers))
	if containers, shouldMutateInit, err = m.MutateContainers(ctx, pod.Spec.InitContainers, pod, config); err != nil {
		return res, errors.Wrapf(err, "MutateContainers Initcontainers ")
	}
	pod.Spec.InitContainers = containers

	config.Log.Debugf("Looking at %d Containers", len(pod.Spec.Containers))
	if containers, shouldMutate, err = m.MutateContainers(ctx, pod.Spec.Containers, pod, config); err != nil {
		return res, errors.Wrapf(err, "MutateContainers containers ")
	}
	pod.Spec.Containers = containers

	if shouldMutate || shouldMutateInit {
		pod.Spec.InitContainers = append(m.CreateInitContainer(config), pod.Spec.InitContainers...)
		pod.Spec.Volumes = append(pod.Spec.Volumes, m.CreateVolume(config)...)

	}
	return &kwhmutating.MutatorResult{MutatedObject: pod}, nil
}

// CreateInitContainer injects an initContainer that copies the asm-env binary to a shared mounted volume
func (m *Mutator) CreateInitContainer(config ASMConfig) []corev1.Container {
	var containers = []corev1.Container{}

	BinPath := fmt.Sprintf("%s%s", config.BinPath, config.BinaryName)
	newPath := fmt.Sprintf("%s%s", config.MountPath, config.BinaryName)
	cmd := fmt.Sprintf("cp %s %s && chmod +x %s", BinPath, newPath, newPath)

	config.Log.Debugf("Appending init container %s", config.ImageName)
	config.Log.Debugf("Command set to %s", cmd)

	containers = append(containers, corev1.Container{
		Name:            "copy-asm-binary",
		Image:           config.ImageName,
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
				Name:      config.BinaryName,
				MountPath: config.MountPath,
			},
		},
	})
	return containers
}

// CreateVolume creates the shared volume that receives asm-env binary
func (m *Mutator) CreateVolume(config ASMConfig) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: config.BinaryName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		},
	}
	return volumes

}
