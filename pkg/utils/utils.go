package utils

import (
	"strings"
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
