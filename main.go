package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	// version and revision are set by goreleaser during the build process.
	version  = "dev"
	revision = "dev"
)

type ContextKey string

const (
	ConfigContextKey ContextKey = "config" // Key for storing loaded config
)

// func main() {
// 	cli.VersionPrinter = func(cmd *cli.Command) {
// 		root := cmd.Root()
// 		fmt.Fprintf(root.Writer, "%s version %s (revision:%s)\n", root.Name, root.Version, revision)
// 	}

// 	app := &cli.Command{
// 		Name:    "ai-rules-manager",
// 		Usage:   "Manage AI agent configuration presets (rules)",
// 		Version: version,
// 		Flags: []cli.Flag{
// 			&cli.StringFlag{
// 				Name:    "config",
// 				Aliases: []string{"c"},
// 				Value:   "ai-presets.toml",
// 				Usage:   "Load configuration from `FILE`",
// 				Sources: cli.EnvVars("AI_PRESETS_CONFIG_LOCATION"),
// 			},
// 		},
// 		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
// 			cfgPath := c.String("config")

// 			loadedCfg, err := config.CreateConfigManager().Load(cfgPath)
// 			if err != nil {
// 				return ctx, fmt.Errorf("failed to load configuration from %s: %w", cfgPath, err)
// 			}

// 			ctxWithConfig := context.WithValue(ctx, ConfigContextKey, loadedCfg)

// 			return ctxWithConfig, nil
// 		},
// 		Action: func(ctx context.Context, cmd *cli.Command) error {
// 			return cli.ShowAppHelp(cmd)
// 		},
// 		Commands: []*cli.Command{
// 			{
// 				Name:   "sync",
// 				Usage:  "Synchronize presets from inputs to outputs based on config",
// 				Action: doSync,
// 			},
// 			{
// 				Name:  "import",
// 				Usage: "Import presets from an existing agent format into the default format",
// 				Flags: []cli.Flag{
// 					&cli.StringFlag{
// 						// Cursor, VSCode GitHub Copilot, etc.
// 						Name:     "from",
// 						Usage:    "Input source to import from",
// 						Aliases:  []string{"f"},
// 						Required: true,
// 					},
// 				},
// 				Action: doImport,
// 			},
// 			{
// 				Name:   "doctor",
// 				Usage:  "Validate preset directory structure or config inputs against the default format",
// 				Action: doDoctor,
// 			},
// 		},
// 	}

// 	if err := app.Run(context.Background(), os.Args); err != nil {
// 		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
// 		os.Exit(1)
// 	}
// }

func main() {

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	matches, err := filepath.Glob(filepath.Join(cwd, ".cursor/rules/**/*.mdc"))
	if err != nil {
		return
	}

	fmt.Println(matches)

	// filepath.WalkDir(".cursor", func(path string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if d.IsDir() {
	// 		return nil
	// 	}

	// 	relPath, err := filepath.Rel(".cursor", path)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	dir := filepath.Base(relPath)

	// 	if !slices.Contains(allowedDirs, dir) {
	// 		return nil
	// 	}

	// 	fmt.Println(relPath)

	// 	return nil
	// })
}
