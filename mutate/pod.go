package mutate

import (
	"context"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *Mutator) MutatePod(ctx context.Context, pod *corev1.Pod) (res *kwhmutating.MutatorResult, err error) {
	// override command
	// override entry point
	var shouldMutate bool

	for i, container := range pod.Spec.Containers {
		var mutatedContainer *corev1.Container

		if shouldMutate, err = m.ContainerHasSecrets(&container, pod.GetNamespace()); err != nil {
			return res, errors.Wrapf(err, "ContainerHasSecrets -  %s", container.Name)
		}
		if !shouldMutate {
			log.Debugf("No asm secrets in container: %s", container.Name)
			continue
		}
		if mutatedContainer, err = m.MutateContainer(ctx, &container, &pod.Spec, pod.GetNamespace()); err != nil {
			log.Debugf("Error mutating container: %s", container.Name)
			return res, errors.Wrapf(err, "MutateContainer - container %s", container.Name)
		}
		pod.Spec.Containers[i] = *mutatedContainer
	}

	if shouldMutate {

	}
	return &kwhmutating.MutatorResult{MutatedObject: pod}, nil
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
