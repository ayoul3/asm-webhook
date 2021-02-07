package mutate

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func (m *Mutator) ContainerHasSecrets(container *corev1.Container, ns string) (hasSecrets bool, err error) {
	for _, env := range container.Env {
		if hasASMPrefix(env.Value) {
			return true, nil
		}
		if env.ValueFrom != nil {
			if hasSecrets, err = m.SourceHasSecret(&env, ns); err != nil {
				return false, errors.Wrapf(err, "lookForValueFrom - %s", env.Name)
			}
			if hasSecrets {
				return true, nil
			}
		}
	}

	if len(container.EnvFrom) > 0 {
		if hasSecrets, err = m.EnvFromHasSecret(&container.EnvFrom, ns); err != nil {
			return false, errors.Wrapf(err, "EnvFromHasSecret - ")
		}
		if hasSecrets {
			return true, nil
		}
	}

	return false, nil
}

func (m *Mutator) MutateContainers(ctx context.Context, containers []corev1.Container, pod *corev1.Pod) (out []corev1.Container, shouldMutate bool, err error) {
	for i, container := range containers {
		var mutatedContainer *corev1.Container

		if shouldMutate, err = m.ContainerHasSecrets(&container, pod.GetNamespace()); err != nil {
			return containers, false, errors.Wrapf(err, "ContainerHasSecrets -  %s", container.Name)
		}
		if !shouldMutate {
			log.Debugf("No asm secrets in container: %s", container.Name)
			continue
		}
		if mutatedContainer, err = m.MutateSingleContainer(ctx, &container, &pod.Spec, pod.GetNamespace()); err != nil {
			log.Debugf("Error mutating container: %s", container.Name)
			return containers, false, errors.Wrapf(err, "MutateSingleContainer - container %s", container.Name)
		}
		containers[i] = *mutatedContainer
	}
	return containers, true, nil
}

func (m *Mutator) MutateSingleContainer(ctx context.Context, container *corev1.Container, podSpec *corev1.PodSpec, ns string) (new *corev1.Container, err error) {
	var imageArgs []string
	args := container.Command
	if len(args) == 0 {
		if imageArgs, err = m.ExtractArgsFromImageConfig(ctx, container, podSpec, ns); err != nil {
			return container, errors.Wrapf(err, "ExtractArgsFromImageConfig - error for %s ", container.Name)
		}
		args = append(args, imageArgs...)
	}

	args = append(args, container.Args...)

	execPath := fmt.Sprintf("%s%s", m.ASMConfig.MountPath, m.ASMConfig.BinaryName)
	container.Command = []string{execPath}
	container.Args = args

	container.VolumeMounts = append(container.VolumeMounts, []corev1.VolumeMount{
		{
			Name:      m.ASMConfig.BinaryName,
			MountPath: m.ASMConfig.MountPath,
		},
	}...)
	return container, nil
}

func (m *Mutator) ExtractArgsFromImageConfig(ctx context.Context, container *corev1.Container, podSpec *corev1.PodSpec, ns string) (args []string, err error) {
	var imageConfig *v1.Config
	args = make([]string, 0)

	if imageConfig, err = m.Registry.GetImageConfig(ctx, m.K8sClient, ns, container, podSpec); err != nil {
		return args, errors.Wrap(err, "GetImageConfig - ")
	}
	log.Debugf("Got image config from Kube: Entrypoint: %s, Cmd: %s", imageConfig.Entrypoint, imageConfig.Cmd)
	args = append(args, imageConfig.Entrypoint...)
	if len(container.Args) == 0 {
		args = append(args, imageConfig.Cmd...)
	}
	return args, nil
}
