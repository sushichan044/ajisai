package fetcher

import (
	"fmt"
)

type InvalidSourceTypeError struct {
	expectedType string
	actualType   string
	err          error
}

func (e *InvalidSourceTypeError) ExpectedType() string {
	return e.expectedType
}

func (e *InvalidSourceTypeError) ActualType() string {
	return e.actualType
}

func (e *InvalidSourceTypeError) Error() string {
	return fmt.Sprintf("expected source type: %s, got: %s", e.expectedType, e.actualType)
}

func (e *InvalidSourceTypeError) Unwrap() error {
	return e.err
}
