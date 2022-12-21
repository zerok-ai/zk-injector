package zkclient

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

func GetCommandFromImage(image string) ([]string, error) {
	dockerClient, _ := client.NewClientWithOpts(client.FromEnv)

	// images, _ := dockerClient.ImageList(context.Background(), types.ImageListOptions{})

	// fmt.Printf("Docker images %v\n", images)

	imageInspect, _, err := dockerClient.ImageInspectWithRaw(context.Background(), "rajeevzerok/zk-injector:0.6")

	if err != nil {
		fmt.Println("Error caught while getting cmd from image: ", image, ", Error is: ", err)
		return []string{}, nil
	}

	return imageInspect.Config.Cmd, nil
}
