package framerUtils

func P[T any](val T) *T {
	return &val
}
