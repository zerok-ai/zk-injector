package inject

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/zerok-ai/zerok-injector/pkg/zkclient"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Injector struct {
	ImageDownloadTracker *zkclient.ImageDownloadTracker
}

func (h *Injector) GetEmptyResponse(admissionReview v1.AdmissionReview) ([]byte, error) {
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

func (h *Injector) Inject(body []byte) ([]byte, error) {
	admissionReview := v1.AdmissionReview{}
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	//var err error
	var pod *corev1.Pod

	responseBody := []byte{}
	ar := admissionReview.Request
	admissionResponse := v1.AdmissionResponse{}
	emptyResponse, _ := h.GetEmptyResponse(admissionReview)

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

		patches, err := h.getPatches(pod)
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

func (h *Injector) getPatches(pod *corev1.Pod) ([]map[string]interface{}, error) {
	p := make([]map[string]interface{}, 0)
	p = append(p, h.getInitContainerPatches(pod)...)
	p = append(p, h.getVolumePatch()...)
	containerPatches, err := h.getContainerPatches(pod)
	if err != nil {
		return make([]map[string]interface{}, 0), err
	}
	p = append(p, containerPatches...)
	fmt.Printf("The patches created are %v.\n", p)
	return p, nil
}

func (h *Injector) getContainerPatches(pod *corev1.Pod) ([]map[string]interface{}, error) {

	p := make([]map[string]interface{}, 0)

	containers := pod.Spec.Containers

	for i := range containers {

		container := &pod.Spec.Containers[i]

		imageType := zkclient.GetImageType(container.Image)

		imageHandler := zkclient.GetImageHandler(imageType)

		podCmd, err := h.getPatchCmdForContainer(container, pod, &imageHandler)

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
			"value": []string{"-c", "/opt/zerok/zerok-agent.sh " + podCmd},
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

func (h *Injector) getPatchCmdForContainer(container *corev1.Container, pod *corev1.Pod, imageHandler *zkclient.ImageHandlerInterface) (string, error) {
	if container == nil {
		fmt.Println("Container is nil.")
		return "", fmt.Errorf("container is nil")
	}
	containerCommand := container.Command
	args := container.Args
	for i := 0; i < len(args); i++ {
		args[i] = strconv.Quote(args[i])
	}
	containerCommand = append(containerCommand, args...)
	var err error
	if len(containerCommand) == 0 {
		containerCommand, err = (*imageHandler).GetCommandFromImage(container.Image, pod, h.ImageDownloadTracker)
	}
	fmt.Println("Container command is ", containerCommand)
	combinedCommand := strings.Join(containerCommand[:], " ")
	combinedCommand = strconv.Quote(combinedCommand)
	if err != nil {
		fmt.Println("Error while getting patch command for image: ", container.Image)
		return "", fmt.Errorf("error while getting patch command for image: %v, erro %v", container.Image, err)
	}
	fmt.Println("Exiting cmd for container ", container.Name, " is ", combinedCommand)
	return combinedCommand, nil
}

func (h *Injector) getVolumePatch() []map[string]interface{} {
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

func (h *Injector) getInitContainerPatches(pod *corev1.Pod) []map[string]interface{} {
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
			Image:           "us-west1-docker.pkg.dev/zerok-dev/stage/init-container:test",
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
