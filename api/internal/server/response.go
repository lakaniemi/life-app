package server

// mapSlice converts each item with toResponse. It always returns a non-nil
// slice, so an empty list encodes as [] rather than null.
func mapSlice[T, U any](items []T, toResponse func(T) U) []U {
	responses := make([]U, len(items))
	for i, item := range items {
		responses[i] = toResponse(item)
	}
	return responses
}
