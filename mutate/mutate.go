package mutate

import (
	"encoding/json"
	"fmt"

	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Mutate mutates
func Mutate(body []byte) (responseBody []byte, err error) {
	var pod *corev1.Pod
	var admResp v1beta1.AdmissionResponse
	var admReview v1beta1.AdmissionReview

	if err = json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}
	ar := admReview.Request
	if ar == nil {
		return responseBody, nil
	}

	// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
	if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
		return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
	}
	// set response options
	admResp.Allowed = true
	admResp.UID = ar.UID
	pT := v1beta1.PatchTypeJSONPatch
	admResp.PatchType = &pT // it's annoying that this needs to be a pointer as you cannot give a pointer to a constant?

	// add some audit annotations, helpful to know why a object was modified, maybe (?)
	admResp.AuditAnnotations = map[string]string{
		"ssm-webhook-resp": "success",
	}

	// the actual mutation is done by a string in JSONPatch style, i.e. we don't _actually_ modify the object, but
	// tell K8S how it should modifiy it
	p := []map[string]string{}
	if len(pod.Spec.Containers) < 1 {
		return responseBody, fmt.Errorf("No containers inside pod request %s", admReview.Request.UID)
	}
	for i := range pod.Spec.Containers {
		patch := map[string]string{
			"op":    "replace",
			"path":  fmt.Sprintf("/spec/containers/%d/image", i),
			"value": "debian",
		}
		p = append(p, patch)
	}
	// parse the []map into JSON
	admResp.Patch, _ = json.Marshal(p)

	admResp.Result = &metav1.Status{
		Status: "Success",
	}

	admReview.Response = &admResp
	// back into JSON so we can return the finished AdmissionReview w/ Response directly
	// w/o needing to convert things in the http handler
	if responseBody, err = json.Marshal(admReview); err != nil {
		return nil, err // untested section
	}

	return responseBody, nil
}
