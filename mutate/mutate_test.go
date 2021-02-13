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
			m.ParseConfig(obj)
			Expect(m.Debug).To(Equal(true))
			Expect(m.ASMConfig.ImageName).To(Equal("special-image:latest"))
			Expect(m.ASMConfig.BinaryName).To(Equal("newBinary"))
			Expect(m.ASMConfig.MountPath).To(Equal("/new/"))
			Expect(m.ASMConfig.BinPath).To(Equal("/bin/"))
		})
	})

})
var _ = Describe("CreateClient", func() {
	Context("When failure to fetch namespace", func() {
		It("should return correct config", func() {
			_, err := mutate.CreateClient(afero.NewMemMapFs())
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When client creation succeeds", func() {
		It("should return no error", func() {
			fs := afero.NewMemMapFs()
			fs.Create("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
			_, err := mutate.CreateClient(fs)
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
