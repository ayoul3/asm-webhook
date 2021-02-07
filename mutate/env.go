package mutate

import (
	"context"
	"strings"

	"emperror.dev/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func hasASMPrefix(name string) bool {
	return strings.Contains(name, "arn:aws:secretsmanager")
}

func (m *Mutator) EnvFromHasSecret(envFrom *[]corev1.EnvFromSource, ns string) (hasSecrets bool, err error) {
	for _, ef := range *envFrom {
		if ef.ConfigMapRef != nil {
			if hasSecrets, err = m.ConfigMapHasSecret(ef.ConfigMapRef.Name, ns, ef.ConfigMapRef.Optional); err != nil {
				return false, errors.Wrapf(err, "ConfigMapHasSecret - %s ", ef.ConfigMapRef.Name)
			}
			if hasSecrets {
				return true, nil
			}
		}
		if ef.SecretRef != nil {
			if hasSecrets, err = m.SecretRefHasSecret(ef.SecretRef.Name, ns, ef.SecretRef.Optional); err != nil {
				return false, errors.Wrapf(err, "SecretRefHasSecret - %s ", ef.SecretRef.Name)
			}
			if hasSecrets {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *Mutator) SourceHasSecret(env *corev1.EnvVar, ns string) (hasSecret bool, err error) {
	var optional bool
	if env.ValueFrom.ConfigMapKeyRef != nil {
		if hasSecret, err = m.ConfigMapHasSecret(env.ValueFrom.ConfigMapKeyRef.Name, ns, &optional); err != nil {
			return false, errors.Wrapf(err, "ConfigMapHasSecret - %s ", env.ValueFrom.ConfigMapKeyRef.Name)
		}
		if hasSecret {
			return true, nil
		}
	}
	if env.ValueFrom.SecretKeyRef != nil {
		if hasSecret, err = m.SecretRefHasSecret(env.ValueFrom.SecretKeyRef.Name, ns, &optional); err != nil {
			return false, errors.Wrapf(err, "SecretRefHasSecret - %s ", env.ValueFrom.SecretKeyRef.Name)
		}
		if hasSecret {
			return true, nil
		}
	}

	return false, nil
}

func (mw *Mutator) ConfigMapHasSecret(cmName string, ns string, optional *bool) (hasSecret bool, err error) {
	configMap, err := mw.K8sClient.CoreV1().ConfigMaps(ns).Get(context.Background(), cmName, v1.GetOptions{})
	if err != nil && (apierrors.IsNotFound(err) || (optional != nil && *optional)) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, value := range configMap.Data {
		if hasASMPrefix(string(value)) {
			return true, nil
		}
	}
	return false, nil
}

func (mw *Mutator) SecretRefHasSecret(secretName string, ns string, optional *bool) (hasSecret bool, err error) {
	secret, err := mw.K8sClient.CoreV1().Secrets(ns).Get(context.Background(), secretName, v1.GetOptions{})

	if err != nil && (apierrors.IsNotFound(err) || (optional != nil && *optional)) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	for _, value := range secret.Data {
		if hasASMPrefix(string(value)) {
			return true, nil
		}
	}
	return false, nil
}
