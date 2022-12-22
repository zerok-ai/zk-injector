package zkclient

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func GetCommandFromImage(image string) ([]string, error) {
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv)

	reader, _ := dockerClient.ImagePull(context.Background(), image, types.ImagePullOptions{})
	defer reader.Close()

	if reader != nil {
		fmt.Println("Pulled the docker image ", image)
	}

	time.Sleep(10 * time.Second)

	imageInspect, _, err := dockerClient.ImageInspectWithRaw(context.Background(), image)

	if err != nil {
		fmt.Println("Error caught while getting cmd from image: ", image, ", Error is: ", err)
		return []string{}, nil
	}

	return imageInspect.Config.Cmd, nil
}
