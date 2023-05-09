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
	RuntimeMapVersion int64
	ImageStore        ImageStore
}

func (h *ImageRuntimeHandler) pollDataFromRedis(redisConfig config.RedisConfig) {
	//Sync first time on pod start
	h.syncDataFromRedis()

	//Creating a timer for periodic sync
	var duration = time.Duration(redisConfig.PollingInterval) * time.Second
	ticker := time.NewTicker(duration)
	for range ticker.C {
		fmt.Println("Sync triggered.")
		h.syncDataFromRedis()
	}
}

func (h *ImageRuntimeHandler) syncDataFromRedis() error {
	versionFromRedis, err := h.ImageStore.GetHashSetVersion()
	if err != nil {
		fmt.Printf("Error caught while getting hash set version from redis %v.\n", err)
		return err
	}
	if h.RuntimeMapVersion == -1 || h.RuntimeMapVersion != versionFromRedis {
		h.RuntimeMapVersion = versionFromRedis
		h.ImageRuntimeMap, err = h.ImageStore.LoadAllData(h.ImageRuntimeMap)
		if err != nil {
			fmt.Printf("Error caught while loading all data from redis %v.\n", err)
			return err
		}
	}
	return nil
}

func (h *ImageRuntimeHandler) Init(redisConfig config.RedisConfig) {
	//init ImageStore
	h.ImageStore = *GetNewImageStore(redisConfig)
	h.RuntimeMapVersion = -1
	h.ImageRuntimeMap = &sync.Map{}
	go h.pollDataFromRedis(redisConfig)
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
	fmt.Printf("Image is %v.\n", imageId)
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
