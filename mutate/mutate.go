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
	K8sClient kubernetes.Interface
	Namespace string
	Registry  registry.ImageRegistry
}

type ASMConfig struct {
	ImageName     string
	MountPath     string
	BinPath       string
	BinaryName    string
	Debug         bool
	SkipCertCheck bool
	Log           *log.Logger
}

func CreateClient(k8sClient kubernetes.Interface, fs afero.Fs) (*Mutator, error) {
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
	}, nil
}

func NewK8SClient() (kubernetes.Interface, error) {
	kubeConfig, err := kubernetesConfig.GetConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeConfig)
}

// SecretsMutator receives the object to mutate and calls the right function according to its type
func (m *Mutator) SecretsMutator(ctx context.Context, _ *kwhmodel.AdmissionReview, obj metav1.Object) (*kwhmutating.MutatorResult, error) {
	asmConfig := m.ParseConfig(obj)

	asmConfig.Log.Debugf("SecretsMutator - Received object %s in namespace %s", obj.GetName(), obj.GetNamespace())
	switch v := obj.(type) {
	case *corev1.Pod:
		asmConfig.Log.Debugf("Mutation request for pod %s", v.GetName())
		return m.MutatePod(ctx, v, asmConfig)

	default:
		return &kwhmutating.MutatorResult{}, nil
	}
}

func (m *Mutator) ParseConfig(obj metav1.Object) ASMConfig {
	annotations := obj.GetAnnotations()
	config := ASMConfig{
		ImageName:     "ayoul3/asm-env",
		MountPath:     "/asm/",
		BinPath:       "/app/",
		BinaryName:    "asm-env",
		Debug:         false,
		SkipCertCheck: false,
		Log:           log.New(),
	}

	if val, _ := annotations["asm.webhook.debug"]; val == "true" {
		config.Log.SetLevel(log.DebugLevel)
		config.Debug = true
	}
	if val, ok := annotations["asm.webhook.asm-env.image"]; ok {
		config.ImageName = val
	}
	if val, ok := annotations["asm.webhook.asm-env.path"]; ok {
		config.BinPath = val
	}
	if val, ok := annotations["asm.webhook.asm-env.bin"]; ok {
		config.BinaryName = val
	}
	if val, ok := annotations["asm.webhook.asm-env.mountPath"]; ok {
		config.MountPath = val
	}
	if val, _ := annotations["asm.webhook.asm-env.skip-cert-check"]; val == "true" {
		config.SkipCertCheck = true
	}
	if !strings.HasSuffix(config.BinPath, "/") {
		config.BinPath = config.BinPath + "/"
	}
	if !strings.HasSuffix(config.MountPath, "/") {
		config.MountPath = config.MountPath + "/"
	}
	return config
}
