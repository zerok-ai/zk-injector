// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package mutate

import (
	"encoding/json"
	"fmt"
	"log"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Mutate mutates
func Mutate(body []byte, verbose bool) ([]byte, error) {
	log.Println("Mutate request received.")
	if verbose {
		log.Printf("recv: %s\n", string(body)) // untested section
	}

	// unmarshal request into AdmissionReview struct
	admReview := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admReview.Request
	resp := v1.AdmissionResponse{}

	if ar != nil {

		// get the Pod object and unmarshal it into its struct, if we cannot, we might as well stop here
		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}
		// set response options
		resp.Allowed = true
		resp.UID = ar.UID
		pT := v1.PatchTypeJSONPatch
		resp.PatchType = &pT // it's annoying that this needs to be a pointer as you cannot give a pointer to a constant?

		// add some audit annotations, helpful to know why a object was modified, maybe (?)
		resp.AuditAnnotations = map[string]string{
			"zk-injector": "yup it did it new",
		}

		// the actual mutation is done by a string in JSONPatch style, i.e. we don't _actually_ modify the object, but
		// tell K8S how it should modifiy it
		p := getPatches()
		// fmt.Printf("The pod name is %v.\n", pod.Name)
		// ///metadata/labels/zk-status
		// for i, container := range pod.Spec.Containers {
		// 	name := container.Name
		// 	patch := map[string]string{
		// 		"op":    "replace",
		// 		"path":  fmt.Sprintf("/spec/containers/%d/name", i),
		// 		"value": name + "-zk-inject",
		// 	}
		// 	p = append(p, patch)
		// }
		// parse the []map into JSON
		resp.Patch, err = json.Marshal(p)

		fmt.Printf("The patches are %v\n", p)

		if err != nil {
			fmt.Printf("Error caught while marshalling the patches %v.\n", err)
		}

		// Success, of course ;)
		resp.Result = &metav1.Status{
			Status: "Success",
		}

		admReview.Response = &resp
		// back into JSON so we can return the finished AdmissionReview w/ Response directly
		// w/o needing to convert things in the http handler
		responseBody, err = json.Marshal(admReview)
		if err != nil {
			return nil, err // untested section
		}
	}

	if verbose {
		log.Printf("resp: %s\n", string(responseBody)) // untested section
	}

	return responseBody, nil
}

func getPatches() []map[string]interface{} {
	p := make([]map[string]interface{}, 0)
	p = append(p, getInitContainerPatches()...)
	p = append(p, getVolumePatch()...)
	p = append(p, getContainerPatches()...)
	return p
}

func getContainerPatches() []map[string]interface{} {
	p := make([]map[string]interface{}, 0)

	addCommand := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/template/spec/containers/0/command",
		"value": []string{"echo", "Rajeev8989", "&&", "sleep", "20000"},
	}

	p = append(p, addCommand)

	// addArgs := map[string]interface{}{
	// 	"op":    "add",
	// 	"path":  "/spec/template/spec/containers/0/args",
	// 	"value": []string{"-c", "/opt/zerok/zerok-agent.sh"},
	// }

	// p = append(p, addArgs)

	addVolumeMount := map[string]interface{}{
		"op":   "add",
		"path": "/spec/template/spec/containers/0/volumeMounts/-",
		"value": corev1.VolumeMount{
			MountPath: "/opt/zerok",
			Name:      "zerok-init",
		},
	}

	p = append(p, addVolumeMount)

	return p
}

func getVolumePatch() []map[string]interface{} {
	p := make([]map[string]interface{}, 0)

	addVolume := map[string]interface{}{
		"op":   "add",
		"path": "/spec/template/spec/volumes/-",
		"value": corev1.Volume{
			Name: "zerok-init",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	p = append(p, addVolume)

	return p
}

func getInitContainerPatches() []map[string]interface{} {
	p := make([]map[string]interface{}, 0)

	//initContainer patches

	initInitialize := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/template/spec/initContainers",
		"value": []corev1.Container{},
	}

	p = append(p, initInitialize)

	container := &corev1.Container{
		Name:            "zerok-init",
		Command:         []string{"cp", "-r", "/opt/zerok/.", "/opt/temp"},
		Image:           "injection-test:0.0.1",
		ImagePullPolicy: corev1.PullNever,
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/opt/temp",
				Name:      "zerok-init",
			},
		},
	}

	addInitContainer := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/template/spec/initContainers/-",
		"value": container,
	}

	p = append(p, addInitContainer)

	return p
}
