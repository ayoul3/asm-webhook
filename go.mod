module github.com/ayoul3/asm-webhook

go 1.15

require (
	emperror.dev/errors v0.8.0
	github.com/google/go-containerregistry v0.4.1-0.20210128200529-19c2b639fab1
	github.com/google/go-containerregistry/pkg/authn/k8schain v0.0.0-20210206001656-4d068fbcb51f
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.3
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.7.0
	github.com/slok/kubewebhook/v2 v2.0.0-beta.2
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.1
)
