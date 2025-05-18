package utils

import "slices"

func ContainsAny[T comparable](source []T, values []T) bool {
	for _, value := range values {
		if slices.Contains(source, value) {
			return true
		}
	}
	return false
}

func RemoveZeroValues[T comparable](source []T) []T {
	filtered := make([]T, 0, len(source))
	zero := getZero[T]()

	for _, value := range source {
		if value != zero {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func getZero[T comparable]() T {
	var zero T
	return zero
}
