package zkclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetCommandFromImage(image string, authConfig *types.AuthConfig) ([]string, error) {

	start := time.Now()
	ctx := context.TODO()
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv)

	var reader io.ReadCloser
	var imagePullOptions types.ImagePullOptions

	if authConfig != nil {
		encodedJSON, err := json.Marshal(&authConfig)
		if err != nil {
			fmt.Println("Error while marshalling Auth details")
			return []string{}, fmt.Errorf("error while marshalling Auth details for image %v, Error is: %v", image, err)
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		imagePullOptions = types.ImagePullOptions{RegistryAuth: authStr}

	} else {
		imagePullOptions = types.ImagePullOptions{}
	}

	reader, err := dockerClient.ImagePull(ctx, image, imagePullOptions)

	if err != nil {
		fmt.Println("Error while pulling the docker image ", err)
		return []string{}, fmt.Errorf("error caught while pulling the image: %v, Error is: %v", image, err)
	}

	io.ReadAll(reader)

	if reader != nil {
		fmt.Println("Pulled the docker image ", image)
	} else {
		return []string{}, fmt.Errorf("image is empty: %v", image)
	}

	defer reader.Close()

	imageInspect, _, err := dockerClient.ImageInspectWithRaw(ctx, image)

	if err != nil {
		fmt.Println("Error caught while getting cmd from image: ", image, ", Error is: ", err)
		return []string{}, fmt.Errorf("error caught while getting cmd from image: %v, Error is: %v", image, err)
	}

	elapsed := time.Since(start)
	fmt.Printf("getting command took %v.\n", int64(elapsed/time.Second))

	return imageInspect.Config.Cmd, nil
}
