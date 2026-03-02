package utils

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// Deref safely dereferences the given pointer, returning its value or the zero value if nil.
func Deref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
