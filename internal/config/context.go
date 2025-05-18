package config

import (
	"context"
	"errors"
)

type contextKey string

const (
	configContextKey contextKey = "config"
)

func StoreInContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configContextKey, cfg)
}

func RetrieveFromContext(ctx context.Context) (*Config, error) {
	cfg, ok := ctx.Value(configContextKey).(*Config)
	if !ok {
		return nil, errors.New("could not retrieve config from context")
	}

	return cfg, nil
}
