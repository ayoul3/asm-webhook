package mutate_test

import (
	"github.com/ayoul3/asm-webhook/mutate"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	fake "k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("ParseConfig", func() {
	Context("When all annotations are present", func() {
		obj := &corev1.Pod{}
		m := mutate.Mutator{
			K8sClient: fake.NewSimpleClientset(),
		}
		obj.SetAnnotations(map[string]string{
			"asm.webhook.debug":             "true",
			"asm.webhook.asm-env.image":     "special-image:latest",
			"asm.webhook.asm-env.path":      "/bin",
			"asm.webhook.asm-env.bin":       "newBinary",
			"asm.webhook.asm-env.mountPath": "/new",
		})
		It("should return correct config", func() {
			config := m.ParseConfig(obj)
			Expect(config.Debug).To(Equal(true))
			Expect(config.ImageName).To(Equal("special-image:latest"))
			Expect(config.BinaryName).To(Equal("newBinary"))
			Expect(config.MountPath).To(Equal("/new/"))
			Expect(config.BinPath).To(Equal("/bin/"))
		})
	})

})
var _ = Describe("CreateClient", func() {
	k8sClient := fake.NewSimpleClientset()
	Context("When failure to fetch namespace", func() {
		It("should return correct config", func() {
			_, err := mutate.CreateClient(k8sClient, afero.NewMemMapFs())
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When client creation succeeds", func() {
		It("should return no error", func() {
			fs := afero.NewMemMapFs()
			fs.Create("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
			_, err := mutate.CreateClient(k8sClient, fs)
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
