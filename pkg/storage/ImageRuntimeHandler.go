package storage

import (
	"fmt"
	"sync"

	"github.com/zerok-ai/zerok-injector/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type ImageRuntimeHandler struct {
	ImageRuntimeMap sync.Map
}

func (h *ImageRuntimeHandler) SaveRuntimeForImage(imageID string, runtimeDetails *common.ContainerRuntime) {
	fmt.Println("Saving image data for ", imageID)
	h.ImageRuntimeMap.Store(imageID, runtimeDetails)
}

func (h *ImageRuntimeHandler) getRuntimeForImage(imageID string) *common.ContainerRuntime {
	value, ok := h.ImageRuntimeMap.Load(imageID)
	if !ok {
		return nil
	}
	switch y := value.(type) {
	case *common.ContainerRuntime:
		return y
	default:
		return nil
	}
}

func (h *ImageRuntimeHandler) GetContainerCommand(container *corev1.Container, pod *corev1.Pod) string {
	imageId := container.Image
	runtime := h.getRuntimeForImage(imageId)
	if runtime == nil {
		return ""
	}
	processes := runtime.Process
	if len(processes) > 0 {
		process := processes[0]
		fmt.Println("found process ", process)
		if process.Runtime == common.JavaProgrammingLanguage {
			fmt.Println("found cmdline ", process.CmdLine)
			return process.CmdLine
		}
	}
	return ""
}
