package mutate_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/ayoul3/ssm-webhook/mutate"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	v1beta1 "k8s.io/api/admission/v1beta1"
)

func TestLib(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "ssm-webhook - Mutate", []Reporter{reporters.NewJUnitReporter("mutate_report-lib.xml")})
}

var _ = Describe("Mutate", func() {
	Context("When the mutation succeed", func() {
		It("should change the image", func() {
			var admReq v1beta1.AdmissionReview
			var admResp v1beta1.AdmissionReview
			request, _ := ioutil.ReadFile("../res/create-pod-v1.json")
			json.Unmarshal(request, &admReq)
			resp, err := mutate.Mutate(request)
			Expect(err).ToNot(HaveOccurred())
			err = json.Unmarshal(resp, &admResp)
			Expect(err).ToNot(HaveOccurred())
			Expect(admResp.Response.UID).To(Equal(admReq.Request.UID))
			Expect(admResp.Response.UID).To(Equal(admReq.Request.UID))
			Expect(admResp.Response.Patch).To(Equal([]byte(`[{"op":"replace","path":"/spec/containers/0/image","value":"debian"}]`)))
		})
	})
	Context("When the mutation fails", func() {
		It("should return failure", func() {
			request, _ := ioutil.ReadFile("../res/create-empty-pod.json")
			_, err := mutate.Mutate(request)
			Expect(err).To(HaveOccurred())
		})
	})
})
