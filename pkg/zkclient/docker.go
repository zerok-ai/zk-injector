package zkclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetCommandFromImage(image string, authConfig *types.AuthConfig) ([]string, error) {
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

	reader, _ = dockerClient.ImagePull(ctx, image, imagePullOptions)

	defer reader.Close()

	io.ReadAll(reader)

	if reader != nil {
		fmt.Println("Pulled the docker image ", image)
	}

	imageInspect, _, err := dockerClient.ImageInspectWithRaw(ctx, image)

	if err != nil {
		fmt.Println("Error caught while getting cmd from image: ", image, ", Error is: ", err)
		return []string{}, fmt.Errorf("error caught while getting cmd from image: %v, Error is: %v", image, err)
	}

	return imageInspect.Config.Cmd, nil
}
