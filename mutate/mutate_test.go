package mutate_test

import (
	"context"
	"testing"

	"github.com/ayoul3/ssm-webhook/mutate"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestLib(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "ssm-webhook - Mutate", []Reporter{reporters.NewJUnitReporter("mutate_report-lib.xml")})
}

var _ = Describe("Mutate", func() {
	Context("When the mutation succeed", func() {
		It("should change the image", func() {
			initialPod := &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Image: "nginx"},
					},
				},
			}
			resp, err := mutate.MutatePod(context.Background(), initialPod)
			Expect(err).ToNot(HaveOccurred())
			v, _ := resp.MutatedObject.(*corev1.Pod)
			Expect(v.Spec.Containers[0].Image).To(Equal("debian"))

		})
	})
	Context("When the mutation fails", func() {
		It("should return failure", func() {
			/*request, _ := ioutil.ReadFile("../res/create-empty-pod.json")
			_, err := mutate.MutatePod(request)
			Expect(err).To(HaveOccurred())*/
		})
	})
})
