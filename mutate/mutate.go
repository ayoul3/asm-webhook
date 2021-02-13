package mutate

import (
	"context"
	"os"
	"strings"

	"emperror.dev/errors"
	"github.com/ayoul3/asm-webhook/registry"
	log "github.com/sirupsen/logrus"
	kwhmodel "github.com/slok/kubewebhook/v2/pkg/model"
	kwhmutating "github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubernetesConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Mutator struct {
	K8sClient  kubernetes.Interface
	Namespace  string
	Registry   registry.ImageRegistry
	MountPath  string
	BinaryName string
	ASMConfig  ASMConfig
	Debug      bool
}

type ASMConfig struct {
	ImageName  string
	MountPath  string
	BinPath    string
	BinaryName string
}

func CreateClient(fs afero.Fs) (*Mutator, error) {
	k8sClient, err := newK8SClient()
	if err != nil {
		return nil, errors.Wrapf(err, "newK8SClient ")
	}
	namespace, err := afero.ReadFile(fs, "/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return nil, errors.Wrapf(err, "ReadFile namespace:  ")
	}
	if os.Getenv("ASM_DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	return &Mutator{
		K8sClient: k8sClient,
		Namespace: string(namespace),
		Registry:  registry.NewRegistry(),
		ASMConfig: ASMConfig{
			ImageName:  "ayoul3/asm-env",
			MountPath:  "/asm/",
			BinPath:    "/app/",
			BinaryName: "asm-env",
		},
	}, nil
}

func newK8SClient() (kubernetes.Interface, error) {
	kubeConfig, err := kubernetesConfig.GetConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeConfig)
}

// SecretsMutator receives the object to mutate and calls the right function according to its type
func (m *Mutator) SecretsMutator(ctx context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	m.ParseConfig(obj)

	log.Debugf("SecretsMutator - Received object %s in namespace %s", obj.GetName(), obj.GetNamespace())
	switch v := obj.(type) {
	case *corev1.Pod:
		log.Debugf("Got pod %s", v.GetName())
		return m.MutatePod(ctx, v)

	default:
		return &kwhmutating.MutatorResult{}, nil
	}
}

func (m *Mutator) ParseConfig(obj metav1.Object) {
	annotations := obj.GetAnnotations()

	if _, ok := annotations["asm.webhook.debug"]; ok {
		log.SetLevel(log.DebugLevel)
		m.Debug = true
	}
	if val, ok := annotations["asm.webhook.asm-env.image"]; ok {
		m.ASMConfig.ImageName = val
	}
	if val, ok := annotations["asm.webhook.asm-env.path"]; ok {
		m.ASMConfig.BinPath = val
	}
	if val, ok := annotations["asm.webhook.asm-env.bin"]; ok {
		m.ASMConfig.BinaryName = val
	}
	if val, ok := annotations["asm.webhook.asm-env.mountPath"]; ok {
		m.ASMConfig.MountPath = val
	}
	if !strings.HasSuffix(m.ASMConfig.BinPath, "/") {
		m.ASMConfig.BinPath = m.ASMConfig.BinPath + "/"
	}
	if !strings.HasSuffix(m.ASMConfig.MountPath, "/") {
		m.ASMConfig.MountPath = m.ASMConfig.MountPath + "/"
	}
}
