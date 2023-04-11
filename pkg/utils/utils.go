package utils

func FindString(array []string, element string) int {
	for i := 0; i < len(array); i++ {
		if array[i] == element {
			return i
		}
	}
	return -1
}

func AppendArray(array []string, elements []string, index int) []string {
	array = append(array, elements...)
	copy(array[index+len(elements):], array[index:])
	k := 0
	for i := index; k < len(elements); i++ {
		array[i] = elements[k]
		k++
	}
	return array
}
