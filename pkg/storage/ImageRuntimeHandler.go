package storage

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zerok-ai/zerok-injector/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type ImageRuntimeHandler struct {
	ImageRuntimeMap sync.Map
}

func (h *ImageRuntimeHandler) SaveRuntimeForImage(imageID string, runtimeDetails *common.ContainerRuntime) {

	//oldRuntimeDetails := h.getRuntimeForImage(imageID)
	//
	//if oldRuntimeDetails != nil {
	//	fmt.Println("mk: old data for ", imageID, "=== ", *ToJsonString(oldRuntimeDetails))
	//}

	// store only iff there is at least one process running in the container
	if len(runtimeDetails.Process) > 0 {
		h.ImageRuntimeMap.Store(imageID, runtimeDetails)
		fmt.Println("Data received for ", imageID, "=== ", *ToJsonString(runtimeDetails))
	}
}

func ToJsonString(iInstance interface{}) *string {
	if iInstance == nil {
		return nil
	}
	bytes, error := json.Marshal(iInstance)
	if error != nil {
		//TODO:Refactor
		return nil
	} else {
		iString := string(bytes)
		return &iString

	}
}

func (h *ImageRuntimeHandler) getRuntimeForImage(imageID string) *common.ContainerRuntime {
	value, ok := h.ImageRuntimeMap.Load(imageID)
	if !ok {
		return nil
	}
	switch y := value.(type) {
	case *common.ContainerRuntime:
		fmt.Println("mk: Getting data for image id ", imageID, *ToJsonString(y))
		return y
	default:
		return nil
	}
}

func (h *ImageRuntimeHandler) GetContainerCommand(container *corev1.Container, pod *corev1.Pod) (string, common.ProgrammingLanguage) {
	imageId := container.Image
	runtime := h.getRuntimeForImage(imageId)
	if runtime == nil {
		return "", common.UknownLanguage
	}
	processes := runtime.Process
	if len(processes) > 0 {
		process := processes[0]
		fmt.Println("found process ", process)
		if process.Runtime == common.JavaProgrammingLanguage {
			fmt.Println("found cmdline ", process.CmdLine)
			return process.CmdLine, common.JavaProgrammingLanguage
		}
	}
	return "", common.UknownLanguage
}
