package zkclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Map will have image string and list of wait groups.
type DockerImageDownloader struct {
	DownloadCompMap sync.Map
}

func (h *DockerImageDownloader) downloadImage(authConfig *types.AuthConfig, dockerClient *client.Client, image string) error {
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
		err := h.pullImage(authConfig, dockerClient, image)
		h.closeWaitGroups(image)
		return err
	}
	return nil
}

func (h *DockerImageDownloader) closeWaitGroups(image string) {
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

func (h *DockerImageDownloader) pullImage(authConfig *types.AuthConfig, dockerClient *client.Client, image string) error {
	fmt.Println("Pull image method called.")
	start := time.Now()
	var reader io.ReadCloser

	var imagePullOptions types.ImagePullOptions
	ctx := context.TODO()

	if authConfig != nil {
		encodedJSON, err := json.Marshal(&authConfig)
		if err != nil {
			fmt.Println("Error while marshalling Auth details")
			return fmt.Errorf("error while marshalling Auth details for image %v, Error is: %v", image, err)
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		imagePullOptions = types.ImagePullOptions{RegistryAuth: authStr}

	} else {
		imagePullOptions = types.ImagePullOptions{}
	}
	reader, err := dockerClient.ImagePull(ctx, image, imagePullOptions)

	if err != nil {
		fmt.Println("Error while pulling the docker image ", err)
		return fmt.Errorf("error caught while pulling the image: %v, Error is: %v", image, err)
	}

	io.ReadAll(reader)

	if reader != nil {
		fmt.Println("Pulled the docker image ", image)
		reader.Close()
	} else {
		return fmt.Errorf("image is empty: %v", image)
	}
	fmt.Println("Pull image method completed.")
	elapsed := time.Since(start)
	fmt.Printf("Pulling image took %v for image %v.\n", int64(elapsed/time.Second), image)

	return nil
}

func GetCommandFromImage(image string, authConfig *types.AuthConfig, h *DockerImageDownloader) ([]string, error) {
	ctx := context.TODO()
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv)

	err := h.downloadImage(authConfig, dockerClient, image)

	if err != nil {
		return []string{}, fmt.Errorf("image is empty: %v", image)
	}

	imageInspect, _, err := dockerClient.ImageInspectWithRaw(ctx, image)

	if err != nil {
		fmt.Println("Error caught while getting cmd from image: ", image, ", Error is: ", err)
		return []string{}, fmt.Errorf("error caught while getting cmd from image: %v, Error is: %v", image, err)
	}

	return imageInspect.Config.Cmd, nil
}
