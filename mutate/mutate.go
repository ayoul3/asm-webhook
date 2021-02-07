package mutate

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/ayoul3/asm-webhook/registry"
	"k8s.io/client-go/kubernetes"
	kubernetesConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Mutator struct {
	K8sClient kubernetes.Interface
	Namespace string
	Registry  registry.ImageRegistry
}

func CreateClient() *Mutator {
	k8sClient, err := newK8SClient()
	if err != nil {
		log.Fatalf("error creating k8s client: %s", err)
	}
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatalf("error reading k8s namespace: %s", err)
	}
	return &Mutator{
		K8sClient: k8sClient,
		Namespace: string(namespace),
		Registry:  registry.NewRegistry(),
	}
}

func newK8SClient() (kubernetes.Interface, error) {
	kubeConfig, err := kubernetesConfig.GetConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubeConfig)
}
