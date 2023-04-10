package utils

func FindString(array []string, element string) int {
	for i := 0; i < len(array); i++ {
		if array[i] == element {
			return i
		}
	}
	return -1
}

func AppendItem(array []string, element string, i int) []string {
	array = append(array, "")
	copy(array[i+1:], array[i:])
	array[i] = element
	return array
}
