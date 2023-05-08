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

func (h *ImageRuntimeHandler) GetContainerLanguage(container *corev1.Container, pod *corev1.Pod) common.ProgrammingLanguage {
	imageId := container.Image
	runtime := h.getRuntimeForImage(imageId)
	if runtime == nil {
		return common.UknownLanguage
	}
	languages := runtime.Languages
	if len(languages) > 0 {
		language := languages[0]
		fmt.Println("found language ", language)
		if language == fmt.Sprintf("%v", common.JavaProgrammingLanguage) {
			return common.JavaProgrammingLanguage
		}
	}
	return common.UknownLanguage
}
