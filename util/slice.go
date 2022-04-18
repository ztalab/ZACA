package util

func StringSliceToInterfaceSlice(data []string) []interface{} {
	result := make([]interface{}, 0, len(data))
	for _, v := range data {
		result = append(result, v)
	}
	return result
}
