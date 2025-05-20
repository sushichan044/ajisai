package config

import (
	"context"
	"errors"
)

type (
	contextKey string

	Context struct {
		Config *Config

		NotFound bool
	}
)

const (
	configContextKey contextKey = "config"
)

func StoreInContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configContextKey, &Context{
		Config: cfg,
	})
}

func StoreNotFoundInContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, configContextKey, &Context{
		NotFound: true,
	})
}

func RetrieveFromContext(ctx context.Context) (*Context, error) {
	cfgCtx, ok := ctx.Value(configContextKey).(*Context)
	if !ok {
		return nil, errors.New("could not retrieve config from context")
	}

	return cfgCtx, nil
}
