package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"zerok-injector/pkg/common"

	"zerok-injector/pkg/zkclient"

	corev1 "k8s.io/api/core/v1"
)

func FindString(array []string, element string) int {
	for i := 0; i < len(array); i++ {
		if strings.Compare(element, array[i]) == 0 {
			return i
		}
	}
	return -1
}

func AppendArray(array []string, elements []string, index int) []string {
	array = append(array, elements...)
	copy(array[index+len(elements):], array[index:])
	k := 0
	for i := index + 1; k < len(elements); i++ {
		array[i] = elements[k]
		k++
	}
	return array
}

func FromJsonString(data string) (*common.ContainerRuntime, error) {
	var runtimeDetails common.ContainerRuntime
	err := json.Unmarshal([]byte(data), &runtimeDetails)
	if err != nil {
		return nil, err
	}
	return &runtimeDetails, nil
}

func ToJsonString(iInstance interface{}) *string {
	if iInstance == nil {
		return nil
	}
	bytes, err := json.Marshal(iInstance)
	if err != nil {
		return nil
	}
	iString := string(bytes)
	return &iString
}

func GetIndexOfEnv(envVars []corev1.EnvVar, targetEnv string) int {
	for index, envVar := range envVars {
		if envVar.Name == targetEnv {
			return index
		}
	}
	return -1
}

func PrintAllNonOrchestratedPods() {
	podlist, err := zkclient.GetAllNonOrchestratedPods()
	fmt.Printf("Getting all pods.\n")

	if err != nil {
		fmt.Printf("Error %v.\n", err)
	} else {
		for _, pod := range podlist {
			fmt.Printf("Pod with name %v.\n", pod.ObjectMeta.Name)
		}
	}
}
