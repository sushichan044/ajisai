package fetcher

import (
	"fmt"

	"github.com/sushichan044/ajisai/internal/domain"
)

type InvalidSourceTypeError struct {
	expectedType domain.InputSourceType
	actualType   domain.InputSourceType
	err          error
}

func (e *InvalidSourceTypeError) ExpectedType() domain.InputSourceType {
	return e.expectedType
}

func (e *InvalidSourceTypeError) ActualType() domain.InputSourceType {
	return e.actualType
}

func (e *InvalidSourceTypeError) Error() string {
	return fmt.Sprintf("expected source type: %s, got: %s", e.expectedType, e.actualType)
}

func (e *InvalidSourceTypeError) Unwrap() error {
	return e.err
}
