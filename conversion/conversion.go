package conversion

func SliceToAnySlice[T any](slice []T) []any {
	ret := make([]interface{}, len(slice))
	for i, s := range slice {
		ret[i] = s
	}
	return ret
}
