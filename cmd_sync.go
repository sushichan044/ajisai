package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sushichan044/ai-rules-manager/internal/config"
	"github.com/sushichan044/ai-rules-manager/internal/engine"
)

func doSync(c context.Context, cmd *cli.Command) error {
	cfg, err := config.RetrieveFromContext(c)
	if err != nil {
		return fmt.Errorf("failed to retrieve config from context: %w", err)
	}

	engine := engine.NewEngine(cfg)

	packageNames, fetchErr := engine.Fetch()
	if fetchErr != nil {
		return fmt.Errorf("failed to fetch inputs: %w", fetchErr)
	}

	presets, parseErr := engine.Parse(packageNames)
	if parseErr != nil {
		return fmt.Errorf("failed to parse presets: %w", parseErr)
	}

	writeErr := engine.Write(presets)
	if writeErr != nil {
		return fmt.Errorf("failed to write presets: %w", writeErr)
	}

	return nil
}
