package zkclient

import (
	"fmt"
	"sync"
)

// Map will have image string and list of wait groups.

type ImageDownloadHandler struct {
	DownloadCompMap sync.Map
}

func (h *ImageDownloadHandler) downloadImage(image string, imageHandler *ImageHandlerInterface) error {
	fmt.Println("Download Image method called for image ", image)

	//TODO: Add a mutex here to avoid multithreading.
	value, ok := h.DownloadCompMap.Load(image)

	if ok {
		fmt.Println("Already download in queue for image ", image)
		var wg sync.WaitGroup
		wg.Add(1)
		switch y := value.(type) {
		case []*sync.WaitGroup:
			y = append(y, &wg)
			h.DownloadCompMap.Store(image, y)
		default:
			panic("Error in image downloadere.")
		}
		wg.Wait()
		fmt.Println("Waiting ended for image ", image)
	} else {
		var a []*sync.WaitGroup
		h.DownloadCompMap.Store(image, a)
		fmt.Println("First Image so pull image initiated.")
		err := (*imageHandler).pullImage(image)
		h.closeWaitGroups(image)
		return err
	}
	return nil
}

func (h *ImageDownloadHandler) closeWaitGroups(image string) {
	fmt.Println("CloseWaitGroups method called.")
	value, ok := h.DownloadCompMap.Load(image)
	if ok {
		switch y := value.(type) {
		case []*sync.WaitGroup:
			for _, wg := range y {
				wg.Done()
			}
		default:
			panic("Error in image downloadere.")
		}
	}
	h.DownloadCompMap.Delete(image)
}
