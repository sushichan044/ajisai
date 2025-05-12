package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/sushichan044/aisync/internal/config"
	"github.com/sushichan044/aisync/internal/engine"
)

var (
	// version and revision are set by goreleaser during the build process.
	version = "dev"
	// noglobals error is suppressed by golangci-lint.
	revision = "dev"
)

type ContextKey string

const (
	ConfigContextKey ContextKey = "config" // Key for storing loaded config
)

func main() {
	// reassign error is suppressed by golangci-lint.
	cli.VersionPrinter = func(cmd *cli.Command) {
		root := cmd.Root()
		fmt.Fprintf(root.Writer, "%s version %s (revision:%s)\n", root.Name, root.Version, revision)
	}

	app := &cli.Command{
		Name:                  "aisync",
		Usage:                 "Manage AI agent configuration presets (rules)",
		EnableShellCompletion: true,
		Version:               version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "aisync.toml",
				Usage:   "Load configuration from `FILE`",
				Sources: cli.EnvVars("AISYNC_CONFIG_LOCATION"),
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			cfgPath := c.String("config")

			loadedCfg, err := config.NewManager().Load(cfgPath)
			if err != nil {
				return ctx, fmt.Errorf("failed to load configuration from %s: %w", cfgPath, err)
			}

			return config.StoreInContext(ctx, loadedCfg), nil
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cli.ShowAppHelp(cmd)
		},
		Commands: []*cli.Command{
			{
				Name:   "apply",
				Usage:  "Apply presets to the agent according to the config",
				Action: doApply,
			},
			{
				Name:  "clean",
				Usage: "Clean the cache",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force clean the cache",
						Value:   false,
					},
				},
				Action: doClean,
			},
			{
				Name:  "import",
				Usage: "Import presets from an existing agent format into the default format",
				Flags: []cli.Flag{
					&cli.StringFlag{
						// Cursor, VSCode GitHub Copilot, etc.
						Name:     "from",
						Usage:    "Input source to import from",
						Required: true,
					},
				},
				Action: func(_ context.Context, _ *cli.Command) error {
					return nil
				},
			},
			{
				Name:  "doctor",
				Usage: "Validate preset directory structure or config inputs against the default format",
				Action: func(_ context.Context, _ *cli.Command) error {
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

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
