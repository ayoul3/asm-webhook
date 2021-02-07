package mutate

import (
	"context"

	log "github.com/sirupsen/logrus"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MutatePod(ctx context.Context, pod *corev1.Pod) (res *kwhmutating.MutatorResult, err error) {
	for i, _ := range pod.Spec.Containers {
		log.Info("replaced image")
		pod.Spec.Containers[i].Image = "debian"
	}
	return &kwhmutating.MutatorResult{MutatedObject: pod}, nil
}

func SecretsMutator(ctx context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	log.Info("here")
	switch v := obj.(type) {
	case *corev1.Pod:
		log.Info("got pod")
		return MutatePod(ctx, v)

	default:
		return &kwhmutating.MutatorResult{}, nil
	}
}
