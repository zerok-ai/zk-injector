package inject

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/zerok-ai/zerok-injector/pkg/zkclient"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetEmptyResponse(admissionReview v1.AdmissionReview) ([]byte, error) {
	ar := admissionReview.Request
	if ar != nil {
		admissionResponse := v1.AdmissionResponse{}
		admissionResponse.UID = ar.UID
		admissionResponse.Allowed = true
		patchType := v1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		patches := make([]map[string]interface{}, 0)
		admissionResponse.Patch, _ = json.Marshal(patches)
		admissionResponse.Result = &metav1.Status{
			Status: "Success",
		}
		admissionReview.Response = &admissionResponse
		responseBody, err := json.Marshal(admissionReview)
		if err != nil {
			return nil, fmt.Errorf("error caught while marshalling response %v", err)
		}
		return responseBody, nil
	}
	return nil, fmt.Errorf("empty admission request")
}

func Inject(body []byte) ([]byte, error) {
	admissionReview := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	//var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admissionReview.Request
	admissionResponse := v1.AdmissionResponse{}
	emptyResponse, _ := GetEmptyResponse(admissionReview)

	if ar != nil {

		if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
			return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
		}

		admissionResponse.UID = ar.UID

		dt := time.Now()
		fmt.Println("Got request with uid ", ar.UID, " at time ", dt.String())
		admissionResponse.Allowed = true

		patchType := v1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType

		patches, err := getPatches(pod, string(ar.UID))
		if err != nil {
			fmt.Printf("Error caught while getting the patches %v.\n", err)
			return emptyResponse, err
		}
		admissionResponse.Patch, err = json.Marshal(patches)

		fmt.Printf("The patches are %v\n", patches)

		if err != nil {
			fmt.Printf("Error caught while marshalling the patches %v.\n", err)
			return emptyResponse, err
		}

		admissionResponse.Result = &metav1.Status{
			Status: "Success",
		}

		admissionReview.Response = &admissionResponse

		responseBody, err = json.Marshal(admissionReview)
		if err != nil {
			return emptyResponse, err
		}
	}

	log.Printf("resp: %s\n", string(responseBody))

	return responseBody, nil
}

func getPatches(pod *corev1.Pod, uid string) ([]map[string]interface{}, error) {
	p := make([]map[string]interface{}, 0)
	p = append(p, getInitContainerPatches(pod)...)
	p = append(p, getVolumePatch()...)
	containerPatches, err := getContainerPatches(pod, uid)
	if err != nil {
		return make([]map[string]interface{}, 0), err
	}
	p = append(p, containerPatches...)
	fmt.Printf("The patches created are %v.\n", p)
	return p, nil
}

func getPatchCmdForContainer(container *corev1.Container, authConfig *types.AuthConfig, uid string) ([]string, error) {
	if container == nil {
		fmt.Println("Container is nil.")
		return []string{}, fmt.Errorf("container is nil")
	}
	existingCmd, err := zkclient.GetCommandFromImage(container.Image, authConfig, uid)
	if err != nil {
		fmt.Println("Error while getting patch command for image: ", container.Image)
		return []string{}, fmt.Errorf("error while getting patch command for image: %v, erro %v", container.Image, err)
	}
	fmt.Println("Exiting cmd for container ", container.Name, " is ", existingCmd)
	return existingCmd, nil
}

func getContainerPatches(pod *corev1.Pod, uid string) ([]map[string]interface{}, error) {

	imagePullSecrets := &pod.Spec.ImagePullSecrets

	var secrets []string = []string{}

	for _, imagePullSecret := range *imagePullSecrets {
		secrets = append(secrets, imagePullSecret.Name)
	}

	p := make([]map[string]interface{}, 0)

	containers := pod.Spec.Containers

	for i := range containers {

		authConfig, err := zkclient.GetAuthDetailsFromSecret(secrets, pod.Namespace, pod.Spec.Containers[i].Image)

		if err != nil {
			fmt.Printf("Error caught while getting auth config %v for container %v.\n", err, i)
			return p, fmt.Errorf("error caught while getting auth config %v", err)

		}

		podCmd, err := getPatchCmdForContainer(&pod.Spec.Containers[i], authConfig, uid)

		if err != nil {
			fmt.Printf("Error caught while getting command %v for container %v.\n", err, i)
			return p, fmt.Errorf("error caught while getting command %v", err)

		}

		addCommand := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/containers/" + strconv.Itoa(i) + "/command",
			"value": []string{"/bin/sh"},
		}

		p = append(p, addCommand)

		addArgs := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/containers/" + strconv.Itoa(i) + "/args",
			"value": []string{"-c", "/opt/zerok/zerok-agent.sh " + strings.Join(podCmd, " ")},
		}

		p = append(p, addArgs)

		addVolumeMount := map[string]interface{}{
			"op":   "add",
			"path": "/spec/containers/" + strconv.Itoa(i) + "/volumeMounts/-",
			"value": corev1.VolumeMount{
				MountPath: "/opt/zerok",
				Name:      "zerok-init",
			},
		}

		p = append(p, addVolumeMount)

	}

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
			Image:           "rajeevzerok/init-container:latest",
			ImagePullPolicy: corev1.PullAlways,
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
