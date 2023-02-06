package zkclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DockerImageHandler struct {
	dockerClient *client.Client
}

func (h *DockerImageHandler) pullImage(image string, pod *corev1.Pod) error {
	fmt.Println("Pull image method called.")
	start := time.Now()
	var reader io.ReadCloser

	var imagePullOptions types.ImagePullOptions
	ctx := context.TODO()
	secrets := GetImagePullSecrets(pod)

	authConfig, err := GetAuthDetailsFromSecret(pod, secrets, image)

	if err != nil {
		fmt.Println(" Error caught while getting auth details for image", image)
	}

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
	reader, err = h.dockerClient.ImagePull(ctx, image, imagePullOptions)

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

func (h *DockerImageHandler) GetCommandFromImage(image string, pod *corev1.Pod, tracker *ImageDownloadTracker) ([]string, error) {
	ctx := context.TODO()

	var inter ImageHandlerInterface = h

	err := tracker.downloadImage(image, pod, &inter)

	if err != nil {
		return []string{}, fmt.Errorf("image is empty: %v", image)
	}

	imageInspect, _, err := h.dockerClient.ImageInspectWithRaw(ctx, image)

	if err != nil {
		fmt.Println("Error caught while getting cmd from image: ", image, ", Error is: ", err)
		return []string{}, fmt.Errorf("error caught while getting cmd from image: %v, Error is: %v", image, err)
	}

	return imageInspect.Config.Cmd, nil
}

func GetAuthDetailsFromSecret(pod *corev1.Pod, secretNames []string, image string) (*types.AuthConfig, error) {
	clientSet := GetK8sClient()
	listOptions := metav1.GetOptions{}
	var authConfig *types.AuthConfig

	for _, name := range secretNames {
		secret, err := clientSet.CoreV1().Secrets(pod.Namespace).Get(context.TODO(), name, listOptions)

		if err != nil {
			fmt.Println("Error caught while getting the secret ", err)
			return nil, fmt.Errorf("error caught while getting the secret %v in namespace %v", name, pod.Namespace)
		}

		dockerConfigBytes := secret.Data[dockerConfigKey]

		dockerConfigMap := make(map[string]interface{})
		err = json.Unmarshal(dockerConfigBytes, &dockerConfigMap)

		if err != nil {
			fmt.Println("Error caught while unmarshalling the secret ", err)
			return nil, fmt.Errorf("error caught while unmarshalling the secret %v in namespace %v", name, pod.Namespace)
		}

		authValuesMap := dockerConfigMap[authsKey].(map[string]interface{})

		//fmt.Printf("Auth values map is %v", authValuesMap)

		for key, value := range authValuesMap {
			if strings.Contains(image, key) {
				//found the values for the image
				fmt.Printf("Value found for key %v\n", key)
				fmt.Printf("Value for key is %v\n", value)
				valueMap := value.(map[string]interface{})
				username, uok := valueMap["username"]
				passwd, passok := valueMap["password"]
				if uok && passok {
					authConfig = &types.AuthConfig{
						Username: username.(string),
						Password: passwd.(string),
					}
				} else {
					auth, ok := valueMap["auth"]
					if ok {
						fmt.Printf("Auth found for key %v\n", key)
						authConfig = &types.AuthConfig{
							Auth: auth.(string),
						}
					}
				}
				break
			}
		}
		if authConfig != nil {
			break
		}
	}

	return authConfig, nil
}
