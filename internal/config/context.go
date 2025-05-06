package config

import (
	"context"
	"errors"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

type ContextKey string

const (
	configContextKey ContextKey = "config"
)

func StoreInContext(ctx context.Context, cfg *domain.Config) context.Context {
	return context.WithValue(ctx, configContextKey, cfg)
}

func RetrieveFromContext(ctx context.Context) (*domain.Config, error) {
	cfg, ok := ctx.Value(configContextKey).(*domain.Config)
	if !ok {
		return nil, errors.New("could not retrieve config from context")
	}

	return cfg, nil
}
