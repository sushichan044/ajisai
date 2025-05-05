package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/sushichan044/ai-rules-manager/internal/config"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/urfave/cli/v3"
)

var (
	// version and revision are set by goreleaser during the build process
	version  = "dev"
	revision = "dev"
)

// contextKey is a private type to avoid context key collisions.
type contextKey string

const (
	loggerKey contextKey = "logger"
	configKey contextKey = "config" // Key for storing loaded config
)

func main() {
	// Custom version printer
	cli.VersionPrinter = func(cmd *cli.Command) {
		root := cmd.Root()
		fmt.Fprintf(root.Writer, "%s version %s (revision:%s)\n", root.Name, root.Version, revision)
	}

	app := &cli.Command{
		Name:    "ai-rules-manager",
		Usage:   "Manage AI agent configuration presets (rules)",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "ai-rules.toml", // Default config file path
				Usage:   "Load configuration from `FILE`",
				Sources: cli.EnvVars("AI_RULES_CONFIG"), // Allow override via env var
			},
			&cli.StringFlag{
				Name:    "log-level",
				Value:   "info",
				Usage:   "Set log level (debug, info, warn, error)",
				Sources: cli.EnvVars("AI_RULES_LOG_LEVEL"),
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) error {
			// Setup Logger
			logLevel := slog.LevelInfo // Default
			switch strings.ToLower(cmd.String("log-level")) {
			case "debug":
				logLevel = slog.LevelDebug
			case "warn":
				logLevel = slog.LevelWarn
			case "error":
				logLevel = slog.LevelError
			case "info":
				// Default, no change needed
			default:
				if cmd.IsSet("log-level") { // Only warn if user explicitly set an unknown value
					slog.Warn(fmt.Sprintf("Unknown log level '%s', using 'info'", cmd.String("log-level")))
				}
			}

			opts := &slog.HandlerOptions{
				Level: logLevel,
			}
			// Using TextHandler for more human-readable logs, could use JSONHandler
			handler := slog.NewTextHandler(cmd.Root().ErrWriter, opts) // Log to stderr by default
			logger := slog.New(handler)
			slog.SetDefault(logger) // Set as default for convenience

			// Store logger in App.Metadata for command actions
			// Initialize Metadata map if it's nil
			if cmd.Root().Metadata == nil {
				cmd.Root().Metadata = make(map[string]any)
			}
			cmd.Root().Metadata[string(loggerKey)] = logger
			slog.Debug("Logger initialized", "level", logLevel.String())

			// Load Configuration
			configPath := cmd.String("config")
			slog.Debug("Attempting to load configuration", "path", configPath)

			// TODO: Select ConfigManager based on file extension later?
			var cfgMgr domain.ConfigManager = config.NewTomlManager()

			loadedCfg, err := cfgMgr.Load(configPath)
			if err != nil {
				// Log the error, but don't return it from Before hook.
				// Commands that require config must check if it exists in Metadata.
				if os.IsNotExist(err) {
					slog.Info("Configuration file not found, proceeding without loaded config.", "path", configPath)
				} else {
					slog.Warn("Failed to load configuration, proceeding without loaded config.", "path", configPath, "error", err)
				}
				// Do not return the error here: return fmt.Errorf("failed to load configuration from %s: %w", configPath, err)
			} else {
				slog.Info("Configuration loaded successfully", "path", configPath)
				// Store loaded config in App.Metadata only if load was successful
				cmd.Root().Metadata[string(configKey)] = loadedCfg
			}

			return nil // Always return nil from Before hook regarding config loading
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Example of accessing logger and config from Metadata
			logger, _ := cmd.Root().Metadata[string(loggerKey)].(*slog.Logger)
			if logger == nil {
				logger = slog.Default() // Fallback
			}
			cfg, cfgOk := cmd.Root().Metadata[string(configKey)].(*domain.Config)
			// cfgOk will be false if config failed to load

			if cfgOk {
				logger.Info("AI Rules Manager - Main Action (Placeholder)", "loaded_config_namespace", cfg.Global.Namespace)
			} else {
				logger.Info("AI Rules Manager - Main Action (Placeholder) - No config loaded")
			}
			cli.ShowAppHelp(cmd)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "sync",
				Usage: "Synchronize presets from inputs to outputs based on config",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					logger, _ := cmd.Root().Metadata[string(loggerKey)].(*slog.Logger)
					cfg, cfgOk := cmd.Root().Metadata[string(configKey)].(*domain.Config)
					if logger == nil {
						logger = slog.Default()
					} // Fallback
					if !cfgOk {
						logger.Error("Configuration not loaded. Sync requires a valid configuration file.")
						return fmt.Errorf("config not loaded")
					}
					logger.Info("Running sync command (Placeholder)", "namespace", cfg.Global.Namespace)
					// TODO: Implement sync logic using Core Engine
					return nil
				},
			},
			{
				Name:  "import",
				Usage: "Import presets from an existing agent format into the default format",
				// TODO: Add flags like --from, --source, --to
				Action: func(ctx context.Context, cmd *cli.Command) error {
					logger, _ := cmd.Root().Metadata[string(loggerKey)].(*slog.Logger)
					if logger == nil {
						logger = slog.Default()
					} // Fallback
					logger.Info("Running import command (Placeholder)")
					// TODO: Implement import logic
					return nil
				},
			},
			{
				Name:  "doctor",
				Usage: "Validate preset directory structure or config inputs against the default format",
				// TODO: Add optional [directory_path] argument
				Action: func(ctx context.Context, cmd *cli.Command) error {
					logger, _ := cmd.Root().Metadata[string(loggerKey)].(*slog.Logger)
					if logger == nil {
						logger = slog.Default()
					} // Fallback
					logger.Info("Running doctor command (Placeholder)")
					// TODO: Implement doctor logic
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		// We generally expect errors to be handled within actions or the Before hook
		// If an error bubbles up here, log it to stderr
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
