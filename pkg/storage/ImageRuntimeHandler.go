package storage

import (
	"sync"

	"github.com/zerok-ai/zerok-injector/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type ImageRuntimeHandler struct {
	ImageRuntimeMap sync.Map
}

func (h *ImageRuntimeHandler) SaveRuntimeForImage(imageID string, runtimeDetails *common.ContainerRuntime) {
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
	containerStatuses := pod.Status.ContainerStatuses
	for _, containerStatus := range containerStatuses {
		if container.Name == containerStatus.Name {
			imageId := containerStatus.ImageID
			runtime := h.getRuntimeForImage(imageId)
			processes := runtime.Process
			if len(processes) > 0 {
				process := processes[0]
				if process.Runtime == common.JavaProgrammingLanguage {
					return process.CmdLine
				}
			}
		}
	}
	return ""
}
