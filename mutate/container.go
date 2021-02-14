package mutate

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	v1 "github.com/google/go-containerregistry/pkg/v1"
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

func (m *Mutator) MutateContainers(ctx context.Context, containers []corev1.Container, pod *corev1.Pod, config ASMConfig) (out []corev1.Container, shouldMutate bool, err error) {
	for i, container := range containers {
		var mutatedContainer *corev1.Container

		if shouldMutate, err = m.ContainerHasSecrets(&container, pod.GetNamespace()); err != nil {
			return containers, false, errors.Wrapf(err, "ContainerHasSecrets -  %s", container.Name)
		}
		if !shouldMutate {
			config.Log.Debugf("No asm secrets in container: %s", container.Name)
			continue
		}
		config.Log.Debugf("Will mutate container %s", container.Name)
		if mutatedContainer, err = m.MutateSingleContainer(ctx, &container, &pod.Spec, config, pod.GetNamespace()); err != nil {
			config.Log.Warnf("Error mutating container: %s", container.Name)
			return containers, false, errors.Wrapf(err, "MutateSingleContainer - container %s", container.Name)
		}
		containers[i] = *mutatedContainer
	}
	return containers, shouldMutate, nil
}

func (m *Mutator) MutateSingleContainer(ctx context.Context, container *corev1.Container, podSpec *corev1.PodSpec, config ASMConfig, ns string) (new *corev1.Container, err error) {
	var imageArgs []string
	args := container.Command
	if len(args) == 0 {
		config.Log.Debugf("No commands found in Container spec. Will fetch it from registry")
		if imageArgs, err = m.ExtractArgsFromImageConfig(ctx, container, podSpec, ns); err != nil {
			return container, errors.Wrapf(err, "ExtractArgsFromImageConfig - error for %s ", container.Name)
		}
		config.Log.Debugf("Got image config from Kube: %s", imageArgs)
		args = append(args, imageArgs...)
	}

	args = append(args, container.Args...)

	execPath := fmt.Sprintf("%s%s", config.MountPath, config.BinaryName)
	container.Command = []string{execPath}
	container.Args = args

	container.VolumeMounts = append(container.VolumeMounts, []corev1.VolumeMount{
		{
			Name:      config.BinaryName,
			MountPath: config.MountPath,
		},
	}...)

	container.Env = append(container.Env, getEnvVarsFroMConfig(config)...)

	return container, nil
}

func (m *Mutator) ExtractArgsFromImageConfig(ctx context.Context, container *corev1.Container, podSpec *corev1.PodSpec, ns string) (args []string, err error) {
	var imageConfig *v1.Config
	args = make([]string, 0)

	if imageConfig, err = m.Registry.GetImageConfig(ctx, m.K8sClient, ns, container, podSpec); err != nil {
		return args, errors.Wrap(err, "GetImageConfig - ")
	}
	args = append(args, imageConfig.Entrypoint...)
	if len(container.Args) == 0 {
		args = append(args, imageConfig.Cmd...)
	}
	return args, nil
}

func getEnvVarsFroMConfig(config ASMConfig) (envs []corev1.EnvVar) {
	if config.Debug {
		envs = append(envs, corev1.EnvVar{
			Name:  "ASM_DEBUG",
			Value: "true",
		})
	}
	if config.SkipCertCheck {
		envs = append(envs, corev1.EnvVar{
			Name:  "ASM_SKIP_SSL",
			Value: "true",
		})
	}
	return envs
}
