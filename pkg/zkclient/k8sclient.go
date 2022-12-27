package zkclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	dockerConfigKey string = ".dockerconfigjson"
	authsKey        string = "auths"
)

func GetAuthDetailsFromSecret(names []string, namespace string, image string) (*types.AuthConfig, error) {
	clientSet := GetK8sClient()
	listOptions := metav1.GetOptions{}
	var authConfig *types.AuthConfig

	for _, name := range names {
		secret, err := clientSet.CoreV1().Secrets(namespace).Get(context.TODO(), name, listOptions)

		if err != nil {
			fmt.Println("Error caught while getting the secret ", err)
			return nil, fmt.Errorf("error caught while getting the secret %v in namespace %v", name, namespace)
		}

		dockerConfigBytes := secret.Data[dockerConfigKey]

		dockerConfigMap := make(map[string]interface{})
		err = json.Unmarshal(dockerConfigBytes, &dockerConfigMap)

		if err != nil {
			fmt.Println("Error caught while unmarshalling the secret ", err)
			return nil, fmt.Errorf("error caught while unmarshalling the secret %v in namespace %v", name, namespace)
		}

		authValuesMap := dockerConfigMap[authsKey].(map[string]interface{})

		for key, value := range authValuesMap {
			if strings.Contains(image, key) {
				//found the values for the image
				valueMap := value.(map[string]interface{})
				username, uok := valueMap["username"]
				passwd, passok := valueMap["password"]
				if uok && passok {
					authConfig = &types.AuthConfig{
						Username: username.(string),
						Password: passwd.(string),
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

func GetK8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
