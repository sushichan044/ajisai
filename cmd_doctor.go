package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sushichan044/ai-rules-manager/internal/config"
)

func doDoctor(c context.Context, cmd *cli.Command) error {
	cfg, err := config.RetrieveFromContext(c)
	if err != nil {
		return fmt.Errorf("failed to retrieve config from context: %w", err)
	}

	fmt.Printf("Config Namespace: %s\n", cfg.Global.Namespace)
	fmt.Printf("Inputs: %+v\n", cfg.Inputs)
	fmt.Printf("Outputs: %+v\n", cfg.Outputs)

	return nil
}
