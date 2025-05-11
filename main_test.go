package main_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BinaryPath stores the path to the compiled test binary.
var BinaryPath string

// TestMain sets up the test environment by building the main binary.
func TestMain(m *testing.M) {
	var err error
	// Attempt to build the main binary
	BinaryPath, err = buildBinary()
	if err != nil {
		os.Exit(1)
	}
	defer cleanupBinary(BinaryPath)

	m.Run()
}

// buildBinary compiles the main package and returns the path to the binary.
func buildBinary() (string, error) {
	tempDir, err := os.MkdirTemp("", "main_test_build_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	outputPath := filepath.Join(tempDir, "ai-rules-manager-test")

	// Build command - build the package in the current directory
	buildCmd := exec.Command("go", "build", "-o", outputPath, ".")
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		// Include build output in the error message for easier debugging
		return "", fmt.Errorf("failed to build main binary (output: %s): %w", string(output), err)
	}

	// Return the path and a cleanup function, or just the path
	// If returning a cleanup function, adjust the caller (TestMain)
	return outputPath, nil
}

// cleanupBinary removes the test binary and its directory.
func cleanupBinary(binPath string) {
	if binPath != "" {
		dir := filepath.Dir(binPath)
		os.RemoveAll(dir) // Remove the temp directory containing the binary
	}
}

// Test helper function remains the same.
func runCliCommand(_ *testing.T, args []string, env map[string]string) (string, string, error) {
	cmd := exec.Command(BinaryPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if env != nil {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// createValidConfig helper function remains the same.
func createValidConfig(t *testing.T) string {
	td := t.TempDir()
	configPath := filepath.Join(td, "ai-rules.toml")
	content := `
[inputs.test]
type = "local"
path = "./rules"

[outputs.test]
target = "cursor"
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)
	return configPath
}

// Test cases remain largely the same, only adjusting assertions as previously discussed

func TestMain_Run_Version(t *testing.T) {
	stdout, stderr, err := runCliCommand(t, []string{"--version"}, nil)
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, "ai-rules-manager version dev (revision:dev)")
	assert.Empty(t, stderr)
}

func TestMain_Run_ConfigLoading(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (env map[string]string, args []string, cleanup func())
		expectExitCode int
		stdoutContains []string
		stderrContains []string
	}{
		{
			name: "no config flag (expect fallback)",
			setup: func() (map[string]string, []string, func()) {
				// Run sync without creating any config file or flag
				args := []string{"doctor"}
				return nil, args, func() {}
			},
			expectExitCode: 0, // Should run with fallback config
			// TODO: Add assertion for warning log when implemented
		},
		{
			name: "non-existent config via flag (expect fallback)",
			setup: func() (map[string]string, []string, func()) {
				// Use a path that definitely doesn't exist
				nonExistentPath := filepath.Join(t.TempDir(), "non-existent-config.toml")
				args := []string{"doctor", "--config", nonExistentPath}
				return nil, args, func() {}
			},
			expectExitCode: 0, // Should run with fallback config as Load doesn't error
		},
		{
			name: "valid config via flag",
			setup: func() (map[string]string, []string, func()) {
				configPath := createValidConfig(t)
				args := []string{"doctor", "--config", configPath}
				return nil, args, func() { os.Remove(configPath) }
			},
			expectExitCode: 0,
		},
		{
			name: "invalid config via flag (parse error)",
			setup: func() (map[string]string, []string, func()) {
				td := t.TempDir()
				invalidConfigPath := filepath.Join(td, "invalid.toml")
				// Write invalid TOML content (missing closing bracket)
				err := os.WriteFile(invalidConfigPath, []byte(`[inputs.bad`), 0644)
				require.NoError(t, err)
				args := []string{"doctor", "--config", invalidConfigPath}
				return nil, args, func() { os.Remove(invalidConfigPath) }
			},
			expectExitCode: 1,
			// Expecting error from Before hook, wrapped by main error handler
			stderrContains: []string{
				"Error: failed to load configuration",
				"invalid.toml",
				"toml:",
			}, // Check for generic TOML error
		},
		{
			name: "valid config via env var",
			setup: func() (map[string]string, []string, func()) {
				configPath := createValidConfig(t)
				env := map[string]string{"AI_PRESETS_CONFIG_LOCATION": configPath}
				args := []string{"doctor"}
				return env, args, func() { os.Remove(configPath) }
			},
			expectExitCode: 0,
		},
		{
			name: "flag overrides env var (valid flag)",
			setup: func() (map[string]string, []string, func()) {
				envConfigPath := createValidConfig(t)  // Env var points to valid
				flagConfigPath := createValidConfig(t) // Flag points to another valid
				env := map[string]string{"AI_PRESETS_CONFIG_LOCATION": envConfigPath}
				args := []string{"doctor", "--config", flagConfigPath}
				cleanup := func() {
					os.Remove(envConfigPath)
					os.Remove(flagConfigPath)
				}
				return env, args, cleanup
			},
			expectExitCode: 0,
		},
		{
			name: "flag overrides env var (invalid flag - parse error)",
			setup: func() (map[string]string, []string, func()) {
				envConfigPath := createValidConfig(t) // Env var points to valid
				td := t.TempDir()
				invalidConfigPath := filepath.Join(td, "invalid-flag.toml")
				// Write invalid TOML content (e.g., incomplete section)
				err := os.WriteFile(invalidConfigPath, []byte(`[global`), 0644)
				require.NoError(t, err)
				env := map[string]string{"AI_PRESETS_CONFIG_LOCATION": envConfigPath}
				args := []string{"doctor", "--config", invalidConfigPath}
				cleanup := func() {
					os.Remove(envConfigPath)
					os.Remove(invalidConfigPath)
				}
				return env, args, cleanup
			},
			expectExitCode: 1,
			stderrContains: []string{
				"Error: failed to load configuration",
				"invalid-flag.toml",
				"toml:",
			}, // Check for generic TOML error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env, args, cleanup := tc.setup()
			defer cleanup()

			stdout, stderr, err := runCliCommand(t, args, env)

			if tc.expectExitCode == 0 {
				require.NoError(t, err, "stdout: %s\nstderr: %s", stdout, stderr)
			} else {
				require.Error(t, err, "stdout: %s\nstderr: %s", stdout, stderr)
				exitErr := &exec.ExitError{}
				if errors.As(err, &exitErr) {
					assert.Equal(t, tc.expectExitCode, exitErr.ExitCode(), "stdout: %s\nstderr: %s", stdout, stderr)
				} else {
					t.Fatalf("Expected an *exec.ExitError, got %T: %v", err, err)
				}
			}

			for _, contain := range tc.stdoutContains {
				assert.Contains(t, stdout, contain)
			}
			for _, contain := range tc.stderrContains {
				assert.Contains(t, stderr, contain)
			}
		})
	}
}
