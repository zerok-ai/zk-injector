package zkclient

import "github.com/docker/docker/client"

func GetImageHandler(imageType ImageType) ImageHandlerInterface {
	switch imageType {
	case Docker:
		client, _ := client.NewClientWithOpts(client.FromEnv)
		return &DockerImageHandler{
			dockerClient: client,
		}
	}
	return nil
}
