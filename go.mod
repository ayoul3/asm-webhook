module github.com/ayoul3/ssm-webhook

go 1.12

require (
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.15.0
	github.com/sirupsen/logrus v1.7.0
	github.com/slok/kubewebhook v0.11.0
	github.com/slok/kubewebhook/v2 v2.0.0-beta.2
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/klog v0.3.1 // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
)
