package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/sushichan044/ajisai/internal/config"
	"github.com/sushichan044/ajisai/internal/engine"
	"github.com/sushichan044/ajisai/utils"
	"github.com/sushichan044/ajisai/version"
)

var (
	// noglobals error is suppressed by golangci-lint.
	revision = "dev"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// reassign error is suppressed by golangci-lint.
	cli.VersionPrinter = func(cmd *cli.Command) {
		root := cmd.Root()
		fmt.Fprintf(root.Writer, "%s version %s (revision:%s)\n", root.Name, root.Version, revision)
	}

	app := &cli.Command{
		Name:    "ajisai",
		Usage:   "Manage AI agent configuration presets",
		Version: version.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
				Sources: cli.EnvVars("AJISAI_CONFIG_LOCATION"),
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			cfgPath := c.String("config")

			manager, err := prepareConfigManager(cfgPath)
			if err != nil {
				return ctx, fmt.Errorf("failed to prepare loading configuration: %w", err)
			}

			loadedCfg, err := manager.Load()
			if err != nil {
				var configFileNotFound *config.NoFileToReadError
				if errors.As(err, &configFileNotFound) {
					return config.StoreNotFoundInContext(ctx), nil
				}

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

	return app.Run(context.Background(), args)
}

func doApply(c context.Context, _ *cli.Command) error {
	cfgCtx, err := config.RetrieveFromContext(c)
	if err != nil {
		return fmt.Errorf("failed to retrieve config from context: %w", err)
	}

	if cfgCtx.NotFound {
		return errors.New("apply command requires an existing config file")
	}

	cfg := cfgCtx.Config
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	cleanErr := eng.CleanOutputs()
	if cleanErr != nil {
		return fmt.Errorf("failed to clean: %w", cleanErr)
	}

	for packageName := range cfg.Workspace.Imports {
		applyErr := eng.ApplyPackage(packageName)
		if applyErr != nil {
			return fmt.Errorf("failed to apply package %s: %w", packageName, applyErr)
		}
	}

	return nil
}

func doClean(c context.Context, cmd *cli.Command) error {
	cfgCtx, err := config.RetrieveFromContext(c)
	if err != nil {
		return fmt.Errorf("failed to retrieve config from context: %w", err)
	}

	if cfgCtx.NotFound {
		return errors.New("clean command requires an existing config file")
	}

	cfg := cfgCtx.Config

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	force := cmd.Bool("force")
	cleanErr := eng.CleanCache(force)
	if cleanErr != nil {
		return fmt.Errorf("failed to clean cache: %w", cleanErr)
	}

	return nil
}

func prepareConfigManager(cfgPath string) (*config.Manager, error) {
	if cfgPath == "" {
		// init with default file since user didn't specify a config file via -c.
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}

		return config.NewDefaultManagerInDir(cwd)
	}

	absPath, err := utils.ResolveAbsPath(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	return config.NewManager(absPath)
}
