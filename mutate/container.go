package mutate

import (
	"context"
	"strings"

	"emperror.dev/errors"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func (m *Mutator) ContainerHasSecrets(container *corev1.Container) bool {
	for _, env := range container.Env {
		if strings.Contains(env.Value, "arn:aws:secretsmanager") {
			return true
		}
	}
	return false
}

func (m *Mutator) MutateContainer(ctx context.Context, container *corev1.Container, podSpec *corev1.PodSpec, ns string) (new *corev1.Container, err error) {
	var imageArgs []string
	args := container.Command
	if len(args) == 0 {
		if imageArgs, err = m.ExtractArgsFromImageConfig(ctx, container, podSpec, ns); err != nil {
			return container, errors.Wrapf(err, "ExtractArgsFromImageConfig - error getting image config of %s", container.Name)
		}
		args = append(args, imageArgs...)
	}

	args = append(args, container.Args...)

	container.Command = []string{"/asm/asm-env"}
	container.Args = args

	container.VolumeMounts = append(container.VolumeMounts, []corev1.VolumeMount{
		{
			Name:      "asm-env",
			MountPath: "/asm/",
		},
	}...)
	return container, nil
}

func (m *Mutator) ExtractArgsFromImageConfig(ctx context.Context, container *corev1.Container, podSpec *corev1.PodSpec, ns string) (args []string, err error) {
	var imageConfig *v1.Config
	args = make([]string, 0)

	if imageConfig, err = m.Registry.GetImageConfig(ctx, m.K8sClient, ns, container, podSpec); err != nil {
		log.Debugf("Got image config from Kube: Entrypoint: %s, Cmd: %s", imageConfig.Entrypoint, imageConfig.Cmd)
		return args, err
	}
	args = append(args, imageConfig.Entrypoint...)
	if len(container.Args) == 0 {
		args = append(args, imageConfig.Cmd...)
	}
	return args, nil
}
