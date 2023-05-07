package storage

import (
	"fmt"
	"sync"
	"time"

	"zerok-injector/internal/config"
	"zerok-injector/pkg/common"
	"zerok-injector/pkg/utils"

	corev1 "k8s.io/api/core/v1"
)

type ImageRuntimeHandler struct {
	ImageRuntimeMap   *sync.Map
	RuntimeMapVersion *string
	ImageStore        ImageStore
}

func (h *ImageRuntimeHandler) syncDataFromRedis(redisConfig config.RedisConfig) {
	var duration = time.Duration(redisConfig.PollingInterval) * time.Second
	ticker := time.NewTicker(duration)
	for range ticker.C {
		fmt.Println("Sync triggered.")
		versionFromRedis, err := h.ImageStore.GetHashSetVersion()
		if err != nil {
			continue
		}
		if h.RuntimeMapVersion == nil || h.RuntimeMapVersion != versionFromRedis {
			h.RuntimeMapVersion = versionFromRedis
			h.ImageStore.LoadAllData(h.ImageRuntimeMap)
		}
	}
}

func (h *ImageRuntimeHandler) Init(redisConfig config.RedisConfig) {
	//	TODO
	//  2. run async process to check whether all the expected pods have code injected
	//	2.1 Auto-restart pods with auto injection enabled

	//init ImageStore
	h.ImageStore = *GetNewImageStore(redisConfig)
	go h.syncDataFromRedis(redisConfig)
}

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