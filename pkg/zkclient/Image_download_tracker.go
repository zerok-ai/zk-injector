package zkclient

import (
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
)

// Map will have image string and list of wait groups.

type ImageDownloadTracker struct {
	DownloadCompMap sync.Map
}

func (h *ImageDownloadTracker) downloadImage(image string, pod *corev1.Pod, imageHandler *ImageHandlerInterface) error {
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
		err := (*imageHandler).pullImage(image, pod)
		h.closeWaitGroups(image)
		return err
	}
	return nil
}

func (h *ImageDownloadTracker) closeWaitGroups(image string) {
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
