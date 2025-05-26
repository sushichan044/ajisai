package config

import (
	"context"
	"errors"
)

type (
	contextKey string

	// Status represents the status of a configuration.
	Status int

	Context struct {
		Config *Config

		// Status represents the state of the configuration.
		Status Status

		// ValidationError holds the validation error if Status is StatusValidationFailed.
		ValidationError error
	}
)

const (
	configContextKey contextKey = "config"
)

const (
	// StatusValid indicates the configuration is valid and ready to use.
	StatusValid Status = iota

	// StatusNotFound indicates the configuration file was not found.
	StatusNotFound

	// StatusValidationFailed indicates the configuration failed validation.
	StatusValidationFailed

	// Add more status types here as needed...
)

func StoreInContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configContextKey, &Context{
		Config: cfg,
		Status: StatusValid,
	})
}

func StoreNotFoundInContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, configContextKey, &Context{
		Status: StatusNotFound,
	})
}

func StoreValidationErrorInContext(ctx context.Context, validationError error) context.Context {
	return context.WithValue(ctx, configContextKey, &Context{
		Status:          StatusValidationFailed,
		ValidationError: validationError,
	})
}

func RetrieveFromContext(ctx context.Context) (*Context, error) {
	cfgCtx, ok := ctx.Value(configContextKey).(*Context)
	if !ok {
		return nil, errors.New("could not retrieve config from context")
	}

	return cfgCtx, nil
}
