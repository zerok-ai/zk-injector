package utils

import (
	"encoding/json"
	"strings"
	"zerok-injector/pkg/common"
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
		//TODO: Is log needed here?
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
