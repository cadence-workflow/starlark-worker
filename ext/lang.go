package ext

// Must function returns the value if err is nil, otherwise it panics.
// Useful for simplifying error handling in situations where errors are unexpected or unrecoverable.
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}
