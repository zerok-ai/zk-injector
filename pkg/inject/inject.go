// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package inject

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/zerok-ai/zerok-injector/pkg/zkclient"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Inject(body []byte) ([]byte, error) {
	log.Printf("recv: %s\n", string(body)) // untested section

	admissionReview := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admissionReview.Request
	admissionResponse := v1.AdmissionResponse{}

	if ar != nil {

		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}

		admissionResponse.UID = ar.UID
		admissionResponse.Allowed = true

		patchType := v1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType

		patches := getPatches(pod)
		admissionResponse.Patch, err = json.Marshal(patches)

		fmt.Printf("The patches are %v\n", patches)

		if err != nil {
			fmt.Printf("Error caught while marshalling the patches %v.\n", err)
		}

		admissionResponse.Result = &metav1.Status{
			Status: "Success",
		}

		admissionReview.Response = &admissionResponse

		responseBody, err = json.Marshal(admissionReview)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("resp: %s\n", string(responseBody))

	return responseBody, nil
}

func getPatches(pod *corev1.Pod) []map[string]interface{} {
	p := make([]map[string]interface{}, 0)
	p = append(p, getInitContainerPatches(pod)...)
	p = append(p, getVolumePatch()...)
	containerPatches, _ := getContainerPatches(pod)
	p = append(p, containerPatches...)
	return p
}

func getPatchCmdForContainer(container *corev1.Container, authConfig *types.AuthConfig) ([]string, error) {
	if container == nil {
		fmt.Println("Container is nil.")
		return []string{}, fmt.Errorf("container is nil")
	}
	existingCmd, err := zkclient.GetCommandFromImage(container.Image, authConfig)
	if err != nil {
		fmt.Println("Error while getting patch command for image: ", container.Image)
		return []string{}, fmt.Errorf("error while getting patch command for image: %v, erro %v", container.Image, err)
	}
	fmt.Println("Exiting cmd for container ", container.Name, " is ", existingCmd)
	return existingCmd, nil
}

func getContainerPatches(pod *corev1.Pod) ([]map[string]interface{}, error) {

	imagePullSecrets := &pod.Spec.ImagePullSecrets

	p := make([]map[string]interface{}, 0)

	var secrets []string = []string{}

	for _, imagePullSecret := range *imagePullSecrets {
		secrets = append(secrets, imagePullSecret.Name)
	}

	authConfig, err := zkclient.GetAuthDetailsFromSecret(secrets, pod.Namespace, pod.Spec.Containers[0].Image)

	if err != nil {
		fmt.Println("Error caught while getting auth config ", err)
		return p, fmt.Errorf("error caught while getting auth config %v", err)

	}

	getPatchCmdForContainer(&pod.Spec.Containers[0], authConfig)

	addCommand := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/containers/0/command",
		"value": []string{"/bin/sh"},
	}

	p = append(p, addCommand)

	addArgs := map[string]interface{}{
		"op":    "add",
		"path":  "/spec/containers/0/args",
		"value": []string{"-c", "/opt/zerok/zerok-agent.sh"},
	}

	p = append(p, addArgs)

	addVolumeMount := map[string]interface{}{
		"op":   "add",
		"path": "/spec/containers/0/volumeMounts/-",
		"value": corev1.VolumeMount{
			MountPath: "/opt/zerok",
			Name:      "zerok-init",
		},
	}

	p = append(p, addVolumeMount)

	return p, nil
}

func getVolumePatch() []map[string]interface{} {
	p := make([]map[string]interface{}, 0)

	addVolume := map[string]interface{}{
		"op":   "add",
		"path": "/spec/volumes/-",
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

func getInitContainerPatches(pod *corev1.Pod) []map[string]interface{} {
	p := make([]map[string]interface{}, 0)

	if pod.Spec.InitContainers == nil {
		initInitialize := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/initContainers",
			"value": []corev1.Container{},
		}

		p = append(p, initInitialize)
	}

	addInitContainer := map[string]interface{}{
		"op":   "add",
		"path": "/spec/initContainers/-",
		"value": &corev1.Container{
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
		},
	}

	p = append(p, addInitContainer)

	return p
}
