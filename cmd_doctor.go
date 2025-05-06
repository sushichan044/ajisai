package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

func doDoctor(c context.Context, cmd *cli.Command) error {
	cfg := c.Value(ConfigContextKey)

	cfgPtr, ok := cfg.(*domain.Config)

	if !ok {
		return fmt.Errorf("failed to retrieve config from context or invalid config type: %T", cfg)
	}

	fmt.Printf("Config Namespace: %s\n", cfgPtr.Global.Namespace)
	fmt.Printf("Inputs: %+v\n", cfgPtr.Inputs)
	fmt.Printf("Outputs: %+v\n", cfgPtr.Outputs)

	return nil
}
