package utils

// NonNilStrings returns a non-nil shallow copy of src so JSON encoding
// emits `[]` rather than `null` for an empty slice. The plain
// `append([]string(nil), src...)` clone idiom keeps the result nil
// when len(src) == 0 because Go's append doesn't allocate when there
// are no elements to append; that nil then JSON-encodes as `null`,
// which violates wire contracts that declare arrays as non-nullable.
//
// Use at the wire boundary (handler mappers) for any string slice
// that ships out as a JSON array. Cheap (one allocation per call) and
// safe (always returns a fresh backing array, so callers can't
// accidentally alias the source).
func NonNilStrings(src []string) []string {
	out := make([]string, len(src))
	copy(out, src)
	return out
}
