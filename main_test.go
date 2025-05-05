package main_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

// Test helper to build the binary.
func buildBinary(t *testing.T) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "test-ai-rules-manager")
	// Build only the main package
	cmd := exec.Command("go", "build", "-o", binPath, "github.com/sushichan044/ai-rules-manager")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build binary: %s", string(output))
	return binPath
}

func TestMain_Run_Help(t *testing.T) {
	binPath := buildBinary(t)

	// Run without arguments (should show help)
	cmd := exec.Command(binPath)
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	assert.NoError(t, err, "Command should exit successfully when showing help")
	assert.Contains(t, output, "AI Rules Manager - Main Action (Placeholder)") // Check placeholder action output
	assert.Contains(t, output, "USAGE:")
	assert.Contains(t, output, "ai-rules-manager [global options] [command [command options]] [arguments...]")
}

func TestMain_Run_Version(t *testing.T) {
	binPath := buildBinary(t)

	// Run with version flag
	cmd := exec.Command(binPath, "--version")
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	require.NoError(t, err, "Command should exit successfully when showing version")
	assert.Contains(t, output, "ai-rules-manager version dev (revision:dev)")
}

func TestMain_Run_Help_ShowsGlobalFlags(t *testing.T) {
	binPath := buildBinary(t)

	// Run with help flag
	cmd := exec.Command(binPath, "--help")
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	require.NoError(t, err, "Command should exit successfully when showing help")
	assert.Contains(t, output, "GLOBAL OPTIONS:")
	assert.Contains(t, output, "--config FILE", "Should show config flag usage")
	assert.Contains(t, output, "--log-level value", "Should show log-level flag usage")
	assert.Contains(t, output, "AI_RULES_CONFIG", "Should mention config env var")
	assert.Contains(t, output, "AI_RULES_LOG_LEVEL", "Should mention log-level env var")
	assert.Contains(t, output, "(default: \"ai-rules.toml\")", "Should show config default")
	assert.Contains(t, output, "(default: \"info\")", "Should show log-level default")
}

// contextKey is a private type copied from main.go for testing.
type contextKey string

const (
	loggerKey contextKey = "logger"
)

// Test helper to create the app definition (mirrors main.go setup).
func setupTestApp() *cli.Command {
	app := &cli.Command{
		Name: "test-app",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config", // Need flags for Before hook to run
				Value: "dummy.toml",
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info", // Default
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) error {
			// Simplified logger setup for testing - doesn't set global default
			logLevel := slog.LevelInfo
			switch strings.ToLower(cmd.String("log-level")) {
			case "debug":
				logLevel = slog.LevelDebug
			case "warn":
				logLevel = slog.LevelWarn
			case "error":
				logLevel = slog.LevelError
			}
			opts := &slog.HandlerOptions{Level: logLevel}
			handler := slog.NewTextHandler(io.Discard, opts) // Discard output in test
			logger := slog.New(handler)
			if cmd.Root().Metadata == nil {
				cmd.Root().Metadata = make(map[string]any)
			}
			cmd.Root().Metadata[string(loggerKey)] = logger
			return nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error { return nil }, // No-op action
	}
	return app
}

func TestMain_BeforeHook_SetsLogger(t *testing.T) {
	tests := []struct {
		name          string
		logLevelFlag  string
		expectedLevel slog.Level
	}{
		{"default level (info)", "info", slog.LevelInfo},
		{"debug level", "debug", slog.LevelDebug},
		{"warn level", "warn", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"case insensitive", "DEBUG", slog.LevelDebug},
		{"unknown level defaults info", "unknown", slog.LevelInfo},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp()

			// Run the app with the specified log level flag
			args := []string{"test-app", "--log-level", tc.logLevelFlag}
			err := app.Run(context.Background(), args)
			require.NoError(t, err) // Before hook should not return error here

			// Check if the logger is in Metadata and has the correct level
			require.NotNil(t, app.Metadata)
			loggerVal, ok := app.Metadata[string(loggerKey)]
			require.True(t, ok, "Logger should be in metadata")
			logger, ok := loggerVal.(*slog.Logger)
			require.True(t, ok, "Value in metadata should be *slog.Logger")
			assert.True(
				t,
				logger.Enabled(context.Background(), tc.expectedLevel),
				"Logger should be enabled for expected level",
			)

			// Check one level below is disabled (except for debug)
			if tc.expectedLevel > slog.LevelDebug {
				assert.False(
					t,
					logger.Enabled(context.Background(), tc.expectedLevel-1),
					"Logger should be disabled for level below expected",
				)
			}
		})
	}
}

// --- Mock ConfigManager for testing ---.
type mockConfigManager struct {
	LoadFunc func(configPath string) (*domain.Config, error)
	SaveFunc func(configPath string, cfg *domain.Config) error
}

func (m *mockConfigManager) Load(configPath string) (*domain.Config, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(configPath)
	}
	return nil, errors.New("LoadFunc not implemented in mock")
}

func (m *mockConfigManager) Save(configPath string, cfg *domain.Config) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(configPath, cfg)
	}
	return errors.New("SaveFunc not implemented in mock")
}

// Ensure mock implements the interface.
var _ domain.ConfigManager = (*mockConfigManager)(nil)

const configKey contextKey = "config"

// Test helper to create app with mock config manager.
func setupTestAppWithMockConfig(mockMgr domain.ConfigManager) *cli.Command {
	app := &cli.Command{
		Name: "test-app",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: "mock-path.toml"},
			&cli.StringFlag{Name: "log-level", Value: "info"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) error {
			// Logger setup (minimal)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			if cmd.Root().Metadata == nil {
				cmd.Root().Metadata = make(map[string]any)
			}
			cmd.Root().Metadata[string(loggerKey)] = logger

			// Config Loading with mock
			configPath := cmd.String("config")
			loadedCfg, err := mockMgr.Load(configPath)
			if err != nil {
				return fmt.Errorf("mock load failed: %w", err) // Return error to stop app
			}
			cmd.Root().Metadata[string(configKey)] = loadedCfg
			return nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error { return nil },
	}
	return app
}

func TestMain_BeforeHook_LoadsConfig(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string // Path expected by mock
		mockLoadErr error
		mockConfig  *domain.Config
		expectError bool
		expectedCfg *domain.Config // Expected config in metadata
	}{
		{
			name:        "load success",
			configPath:  "correct/path.toml",
			mockConfig:  &domain.Config{Global: domain.GlobalConfig{Namespace: "loaded"}},
			expectError: false,
			expectedCfg: &domain.Config{Global: domain.GlobalConfig{Namespace: "loaded"}},
		},
		{
			name:        "load file not found error",
			configPath:  "notfound.toml",
			mockLoadErr: os.ErrNotExist,
			expectError: true,
		},
		{
			name:        "load other error",
			configPath:  "other/error.toml",
			mockLoadErr: errors.New("some parsing error"),
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockMgr := &mockConfigManager{
				LoadFunc: func(path string) (*domain.Config, error) {
					assert.Equal(t, tc.configPath, path, "Load called with wrong path")
					return tc.mockConfig, tc.mockLoadErr
				},
			}
			app := setupTestAppWithMockConfig(mockMgr)

			args := []string{"test-app", "--config", tc.configPath}
			err := app.Run(context.Background(), args)

			if tc.expectError {
				require.Error(t, err)
				if tc.mockLoadErr != nil {
					assert.ErrorContains(t, err, tc.mockLoadErr.Error())
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, app.Metadata)
				cfgVal, ok := app.Metadata[string(configKey)]
				require.True(t, ok, "Config should be in metadata")
				cfg, ok := cfgVal.(*domain.Config)
				require.True(t, ok, "Value should be *domain.Config")
				assert.Equal(t, tc.expectedCfg, cfg, "Loaded config mismatch")
			}
		})
	}
}

func TestMain_Run_Help_ShowsCommands(t *testing.T) {
	binPath := buildBinary(t)

	// Run with help flag
	cmd := exec.Command(binPath, "--help")
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	require.NoError(t, err, "Command should exit successfully when showing help")
	assert.Contains(t, output, "COMMANDS:")
	assert.Regexp(t, `sync\s+Synchronize presets`, output)
	assert.Regexp(t, `import\s+Import presets`, output)
	assert.Regexp(t, `doctor\s+Validate preset`, output)
	assert.Contains(t, output, "help, h", "Should show help command")
}
