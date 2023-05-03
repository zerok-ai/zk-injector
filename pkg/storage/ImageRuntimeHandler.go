package storage

import (
	"fmt"
	"sync"

	"zerok-injector/internal/config"
	"zerok-injector/pkg/common"
	"zerok-injector/pkg/utils"

	corev1 "k8s.io/api/core/v1"
)

type ImageRuntimeHandler struct {
	ImageRuntimeMap *sync.Map
	ImageStore      ImageStore
}

func (h *ImageRuntimeHandler) Init(redisConfig config.RedisConfig) {
	//	TODO
	//	1. load ImageRuntimeMap through storage.ImageStore
	//  2. run async process to check whether all the expected pods have code injected
	//	2.1 Auto-restart pods with auto injection enabled

	//init ImageStore
	h.ImageStore = *GetNewImageStore(redisConfig)
	h.ImageStore.LoadAllData(h.ImageRuntimeMap)
}

// TODO: Add error handling here. Incase saving to redis fails.
func (h *ImageRuntimeHandler) SaveRuntimeForImage(imageID string, runtimeDetails *common.ContainerRuntime) {

	// store only iff there is at least one process running in the container
	if len(runtimeDetails.Process) > 0 {

		//TODO check if the old value is different from new value then
		// 1. update ImageRuntimeMap
		// 2. storage.ImageStore

		h.ImageRuntimeMap.Store(imageID, runtimeDetails)
		fmt.Println("Data received for ", imageID, "=== ", utils.ToJsonString(runtimeDetails))

	}
}

// TODO: Add error handling here. Incase returning from redis fails.
func (h *ImageRuntimeHandler) getRuntimeForImage(imageID string) *common.ContainerRuntime {
	value, ok := h.ImageRuntimeMap.Load(imageID)
	if !ok {
		return nil
	}
	switch y := value.(type) {
	case *common.ContainerRuntime:
		fmt.Println("mk: Getting data for image id ", imageID, utils.ToJsonString(y))
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
