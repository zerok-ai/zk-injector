package inject

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	utils "github.com/zerok-ai/zerok-injector/pkg/utils"
	"github.com/zerok-ai/zerok-injector/pkg/zkclient"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var agent_options = []string{"-javaagent:/opt/zerok/opentelemetry-javaagent.jar", "-Dotel.javaagent.extensions=/opt/zerok/zk-otel-extension.jar", "-Dotel.traces.exporter=logging"}

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

		podCmd, args, err := h.getCmdAndArgsForContainer(container, pod, &imageHandler)

		if err != nil {
			fmt.Printf("Error caught while getting command %v for container %v.\n", err, i)
			return p, fmt.Errorf("error caught while getting command %v", err)

		}

		podCmd, args = transformCommandAndArgsK8s(podCmd, args)

		fmt.Println("Transformed command ", podCmd, " args ", args)

		addCommand := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/containers/" + strconv.Itoa(i) + "/command",
			"value": podCmd,
		}

		fmt.Println("Add command ", addCommand)

		p = append(p, addCommand)

		addArgs := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/containers/" + strconv.Itoa(i) + "/args",
			"value": args,
		}

		p = append(p, addArgs)

		fmt.Println("Add args ", addArgs)

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

func (h *Injector) getCmdAndArgsForContainer(container *corev1.Container, pod *corev1.Pod, imageHandler *zkclient.ImageHandlerInterface) ([]string, []string, error) {
	if container == nil {
		fmt.Println("Container is nil.")
		return []string{}, []string{}, fmt.Errorf("container is nil")
	}
	containerCommand := container.Command
	args := container.Args
	var err error
	if len(containerCommand) == 0 {
		containerCommand, err = (*imageHandler).GetCommandFromImage(container.Image, pod, h.ImageDownloadTracker)
		args = []string{}
	}
	fmt.Println("Container command ", containerCommand, " and args are ", args)
	if err != nil {
		fmt.Println("Error while getting patch command for image: ", container.Image)
		return []string{}, []string{}, fmt.Errorf("error while getting patch command for image: %v, erro %v", container.Image, err)
	}
	return containerCommand, args, nil
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

func transformCommandAndArgsK8s(command, args []string) ([]string, []string) {
	index := utils.FindString(command, "java")
	if index >= 0 {
		command = utils.AppendArray(command, agent_options, index+1)
	} else {
		index = utils.FindString(args, "java")
		if index >= 0 {
			args = utils.AppendArray(args, agent_options, index+1)
		}
	}
	return command, args
}
