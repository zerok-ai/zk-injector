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

	//TODO: Add a mutex here to avoid multithreading.
	value, ok := h.DownloadCompMap.Load(image)

	if ok {
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
	} else {
		var a []*sync.WaitGroup
		h.DownloadCompMap.Store(image, a)
		err := h.pullImage(authConfig, dockerClient, image)
		h.closeWaitGroups(image)
		return err
	}
	return nil
}

func (h *DockerImageDownloader) closeWaitGroups(image string) {
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
	var reader io.ReadCloser
	defer reader.Close()
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
	} else {
		return fmt.Errorf("image is empty: %v", image)
	}
	return nil
}

func GetCommandFromImage(image string, authConfig *types.AuthConfig, h *DockerImageDownloader) ([]string, error) {

	fmt.Println("New code is running for docker download.")

	start := time.Now()
	fmt.Println("Started pulling the docker image ", image, " at time ", start.String())
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

	elapsed := time.Since(start)
	fmt.Printf("getting command took %v for request.\n", int64(elapsed/time.Second))

	return imageInspect.Config.Cmd, nil
}
