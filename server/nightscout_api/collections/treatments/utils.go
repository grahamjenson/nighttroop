package treatments

func ptr[T any](v T) *T {
	return &v
}
