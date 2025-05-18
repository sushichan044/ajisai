package fetcher

import (
	"fmt"

	"github.com/sushichan044/ajisai/internal/config"
)

type InvalidSourceTypeError struct {
	expectedType config.ImportType
	actualType   config.ImportType
	err          error
}

func (e *InvalidSourceTypeError) ExpectedType() config.ImportType {
	return e.expectedType
}

func (e *InvalidSourceTypeError) ActualType() config.ImportType {
	return e.actualType
}

func (e *InvalidSourceTypeError) Error() string {
	return fmt.Sprintf("expected source type: %s, got: %s", e.expectedType, e.actualType)
}

func (e *InvalidSourceTypeError) Unwrap() error {
	return e.err
}
