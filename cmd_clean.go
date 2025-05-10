package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sushichan044/ai-rules-manager/internal/config"
	"github.com/sushichan044/ai-rules-manager/internal/engine"
)

func doClean(c context.Context, cmd *cli.Command) error {
	cfg, err := config.RetrieveFromContext(c)
	if err != nil {
		return fmt.Errorf("failed to retrieve config from context: %w", err)
	}

	engine, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	force := cmd.Bool("force")
	cleanErr := engine.CleanCache(force)
	if cleanErr != nil {
		return fmt.Errorf("failed to clean cache: %w", cleanErr)
	}

	return nil
}
