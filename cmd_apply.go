package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sushichan044/ai-rules-manager/internal/config"
	"github.com/sushichan044/ai-rules-manager/internal/engine"
)

func doApply(c context.Context, _ *cli.Command) error {
	cfg, err := config.RetrieveFromContext(c)
	if err != nil {
		return fmt.Errorf("failed to retrieve config from context: %w", err)
	}

	engine, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	cleanErr := engine.CleanOutputs()
	if cleanErr != nil {
		return fmt.Errorf("failed to clean: %w", cleanErr)
	}

	packageNames, fetchErr := engine.Fetch()
	if fetchErr != nil {
		return fmt.Errorf("failed to fetch inputs: %w", fetchErr)
	}

	presets, parseErr := engine.Parse(packageNames)
	if parseErr != nil {
		return fmt.Errorf("failed to parse presets: %w", parseErr)
	}

	exportErr := engine.Export(presets)
	if exportErr != nil {
		return fmt.Errorf("failed to export presets: %w", exportErr)
	}

	return nil
}
