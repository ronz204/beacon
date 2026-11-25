package kernel

func IsMissing[T comparable](value T) bool {
	var zero T
	return value == zero
}
