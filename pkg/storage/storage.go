package storage

import (
	"sync"
)

type ImageRuntimeStorage struct {
	ImageRuntimeMap sync.Map
}

type RuntimeDetails struct {
	runtime string
	cmdLine string
}

func (h *ImageRuntimeStorage) saveRuntimeForImage(imageID string, runtimeDetails *RuntimeDetails) {
	h.ImageRuntimeMap.Store(imageID, runtimeDetails)
}

func (h *ImageRuntimeStorage) getRuntimeForImage(imageID string) *RuntimeDetails {
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
