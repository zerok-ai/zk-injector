package zkclient

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetCommandFromImage(image string) ([]string, error) {
	ctx := context.TODO()
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv)

	reader, _ := dockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
	defer reader.Close()

	io.ReadAll(reader)

	if reader != nil {
		fmt.Println("Pulled the docker image ", image)
	}

	imageInspect, _, err := dockerClient.ImageInspectWithRaw(ctx, image)

	if err != nil {
		fmt.Println("Error caught while getting cmd from image: ", image, ", Error is: ", err)
		return []string{}, nil
	}

	return imageInspect.Config.Cmd, nil
}
