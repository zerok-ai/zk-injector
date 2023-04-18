package storage

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
)

type ProgrammingLanguage string

const (
	JavaProgrammingLanguage       ProgrammingLanguage = "java"
	PythonProgrammingLanguage     ProgrammingLanguage = "python"
	GoProgrammingLanguage         ProgrammingLanguage = "go"
	DotNetProgrammingLanguage     ProgrammingLanguage = "dotnet"
	JavascriptProgrammingLanguage ProgrammingLanguage = "javascript"
)

type ImageRuntimeHandler struct {
	ImageRuntimeMap sync.Map
}

type RuntimeDetails struct {
	Runtime ProgrammingLanguage
	CmdLine string
}

func (h *ImageRuntimeHandler) saveRuntimeForImage(imageID string, runtimeDetails *RuntimeDetails) {
	h.ImageRuntimeMap.Store(imageID, runtimeDetails)
}

func (h *ImageRuntimeHandler) getRuntimeForImage(imageID string) *RuntimeDetails {
	value, ok := h.ImageRuntimeMap.Load(imageID)
	if !ok {
		return nil
	}
	switch y := value.(type) {
	case *RuntimeDetails:
		return y
	default:
		return nil
	}
}

func (h *ImageRuntimeHandler) GetContainerCommand(container *corev1.Container, pod *corev1.Pod) *RuntimeDetails {
	containerStatuses := pod.Status.ContainerStatuses
	for _, containerStatus := range containerStatuses {
		if container.Name == containerStatus.Name {
			imageId := containerStatus.ImageID
			return h.getRuntimeForImage(imageId)
		}
	}
	return nil
}
