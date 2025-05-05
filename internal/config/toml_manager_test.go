package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sushichan044/ai-rules-manager/internal/config"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

// Helper function from Validation test
func absPath(t *testing.T, relative string, baseDir string) string {
	abs, err := filepath.Abs(filepath.Join(baseDir, relative))
	require.NoError(t, err)
	return abs
}

func TestTomlManager_Load_Success(t *testing.T) {
	homeDir, _ := os.UserHomeDir() // For ~ expansion test

	tests := []struct {
		name       string
		tomlData   string
		expectedFn func(cfg *domain.Config, configDir string)
	}{
		{
			name: "minimal valid config",
			tomlData: `
[inputs.local_rules]
type = "local"
path = "./rules"

[outputs.cursor]
target = "cursor"
`, // Enabled defaults to true implicitly
			expectedFn: func(cfg *domain.Config, configDir string) {
				assert.Equal(t, "default", cfg.Global.Namespace)
				assert.Equal(t, absPath(t, "./.cache/ai-rules-manager", configDir), cfg.Global.CacheDir)
				require.Len(t, cfg.Inputs, 1)
				input, ok := cfg.Inputs["local_rules"]
				require.True(t, ok)
				assert.Equal(t, "local", input.Type)
				details, ok := input.Details.(domain.LocalInputSourceDetails)
				require.True(t, ok, "Details should be LocalInputSourceDetails")
				assert.Equal(t, absPath(t, "./rules", configDir), details.Path)

				require.Len(t, cfg.Outputs, 1)
				output, ok := cfg.Outputs["cursor"]
				require.True(t, ok)
				assert.Equal(t, "cursor", output.Target)
				assert.True(t, output.Enabled) // Check default is true
			},
		},
		{
			name: "full config with global and git",
			tomlData: `
[global]
cacheDir = "~/.cache/ai-rules"
namespace = "my-proj"

[inputs.remote_rules]
type = "git"
repository = "https://example.com/repo.git"
revision = "main"
subDir = "presets"

[outputs.vscode]
target = "vscode-copilot"
enabled = false
`,
			expectedFn: func(cfg *domain.Config, configDir string) {
				assert.Equal(t, "my-proj", cfg.Global.Namespace)
				if homeDir != "" {
					assert.Equal(t, filepath.Join(homeDir, ".cache/ai-rules"), cfg.Global.CacheDir)
				} else {
					assert.Equal(t, absPath(t, "~/.cache/ai-rules", configDir), cfg.Global.CacheDir)
				}

				require.Len(t, cfg.Inputs, 1)
				input, ok := cfg.Inputs["remote_rules"]
				require.True(t, ok)
				assert.Equal(t, "git", input.Type)
				details, ok := input.Details.(domain.GitInputSourceDetails)
				require.True(t, ok, "Details should be GitInputSourceDetails")
				assert.Equal(t, "https://example.com/repo.git", details.Repository)
				assert.Equal(t, "main", details.Revision)
				assert.Equal(t, "presets", details.SubDir)

				require.Len(t, cfg.Outputs, 1)
				output, ok := cfg.Outputs["vscode"]
				require.True(t, ok)
				assert.Equal(t, "vscode-copilot", output.Target)
				assert.False(t, output.Enabled)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mgr := config.NewTomlManager()

			// Create a temporary file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "ai-rules.toml")
			err := os.WriteFile(configPath, []byte(tc.tomlData), 0644)
			require.NoError(t, err)

			// --- Act
			loadedCfg, err := mgr.Load(configPath) // Call the actual Load method

			// --- Assert
			require.NoError(t, err) // Expect no error for success cases
			require.NotNil(t, loadedCfg)
			if tc.expectedFn != nil {
				// Pass the loaded domain.Config and configDir to the assertion function
				tc.expectedFn(loadedCfg, tempDir)
			}
		})
	}
}

func TestTomlManager_Load_FileNotFound(t *testing.T) {
	mgr := config.NewTomlManager()
	configPath := filepath.Join(t.TempDir(), "non_existent_file.toml")

	// --- Act
	_, err := mgr.Load(configPath)

	// --- Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "config file not found")
	assert.ErrorIs(t, err, os.ErrNotExist) // Check underlying error
}

func TestTomlManager_Load_InvalidToml(t *testing.T) {
	mgr := config.NewTomlManager()
	invalidToml := `
[inputs.local
type = "local" # Missing closing bracket
path = "./rules"
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.toml")
	err := os.WriteFile(configPath, []byte(invalidToml), 0644)
	require.NoError(t, err)

	// --- Act
	_, err = mgr.Load(configPath)

	// --- Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to unmarshal TOML")
	// We might check for a specific toml parsing error type if available/needed,
	// but checking the wrapped message is usually sufficient.
}

func TestTomlManager_Load_ValidationAndDefaults(t *testing.T) {
	// Use helper to get absolute path for comparison
	absPath := func(t *testing.T, relative string, baseDir string) string {
		abs, err := filepath.Abs(filepath.Join(baseDir, relative))
		require.NoError(t, err)
		return abs
	}

	homeDir, _ := os.UserHomeDir() // For ~ expansion test

	tests := []struct {
		name          string
		tomlData      string
		expectError   bool
		errorContains string
		expectedFn    func(cfg *domain.Config, configDir string)
	}{
		{
			name: "valid minimal config - check defaults",
			tomlData: `
[inputs.local1]
type = "local"
path = "./rules"
[outputs.cursor1]
target = "cursor"
`, // enabled defaults true
			expectError: false,
			expectedFn: func(cfg *domain.Config, configDir string) {
				assert.Equal(t, "default", cfg.Global.Namespace)
				assert.Equal(t, absPath(t, "./.cache/ai-rules-manager", configDir), cfg.Global.CacheDir)
				require.Len(t, cfg.Inputs, 1)
				input, ok := cfg.Inputs["local1"]
				require.True(t, ok)
				assert.Equal(t, "local", input.Type)
				details, ok := input.Details.(domain.LocalInputSourceDetails)
				require.True(t, ok)
				assert.Equal(t, absPath(t, "./rules", configDir), details.Path)
				require.Len(t, cfg.Outputs, 1)
				output, ok := cfg.Outputs["cursor1"]
				require.True(t, ok)
				assert.Equal(t, "cursor", output.Target)
				assert.True(t, output.Enabled)
			},
		},
		{
			name: "global overrides and path resolutions",
			tomlData: `
[global]
cacheDir = "../cache/global"
namespace = "override"
[inputs.local_abs]
type = "local"
path = "/abs/path/rules"
[inputs.git1]
type = "git"
repository = "http://a.b/c.git"
[outputs.out1]
target = "target1"
enabled = false
`,
			expectError: false,
			expectedFn: func(cfg *domain.Config, configDir string) {
				assert.Equal(t, "override", cfg.Global.Namespace)
				assert.Equal(t, absPath(t, "../cache/global", configDir), cfg.Global.CacheDir)
				require.Len(t, cfg.Inputs, 2)
				inAbs, _ := cfg.Inputs["local_abs"]
				detailsAbs, _ := inAbs.Details.(domain.LocalInputSourceDetails)
				assert.Equal(t, filepath.Clean("/abs/path/rules"), detailsAbs.Path)
				inGit, _ := cfg.Inputs["git1"]
				detailsGit, _ := inGit.Details.(domain.GitInputSourceDetails)
				assert.Equal(t, "http://a.b/c.git", detailsGit.Repository)
				require.Len(t, cfg.Outputs, 1)
				out, _ := cfg.Outputs["out1"]
				assert.Equal(t, "target1", out.Target)
				assert.False(t, out.Enabled)
			},
		},
		{
			name: "cache dir with home expansion",
			tomlData: `
[global]
cacheDir = "~/mycache"
`,
			expectError: false,
			expectedFn: func(cfg *domain.Config, configDir string) {
				if homeDir != "" {
					assert.Equal(t, filepath.Join(homeDir, "mycache"), cfg.Global.CacheDir)
				} else {
					// If home dir cannot be determined, it should resolve relative to config
					assert.Equal(t, absPath(t, "~/mycache", configDir), cfg.Global.CacheDir)
				}
			},
		},
		{
			name:          "input missing type",
			tomlData:      `[inputs.bad]`, // path = "./rules"`, // Missing type
			expectError:   true,
			errorContains: "missing required 'type' field",
		},
		{
			name: "input unsupported type",
			tomlData: `[inputs.bad]
type="unknown"`,
			expectError:   true,
			errorContains: "unsupported type 'unknown'",
		},
		{
			name: "local input missing path",
			tomlData: `[inputs.bad]
type="local"`,
			expectError:   true,
			errorContains: "type 'local' requires 'path' field",
		},
		{
			name: "local input with git fields",
			tomlData: `
[inputs.bad]
type="local"
path="./p"
repository="a.git"
`,
			expectError:   true,
			errorContains: "type 'local' does not support 'repository'",
		},
		{
			name: "git input missing repository",
			tomlData: `[inputs.bad]
type="git"`,
			expectError:   true,
			errorContains: "type 'git' requires 'repository' field",
		},
		{
			name: "git input with local field",
			tomlData: `
[inputs.bad]
type="git"
repository="a.git"
path="./p"
`,
			expectError:   true,
			errorContains: "type 'git' does not support 'path' field",
		},
		{
			name:          "output missing target",
			tomlData:      `[outputs.bad]`, // enabled=true`, // Missing target
			expectError:   true,
			errorContains: "missing required 'target' field",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mgr := config.NewTomlManager()
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "test-config.toml")
			err := os.WriteFile(configPath, []byte(tc.tomlData), 0644)
			require.NoError(t, err)

			// --- Act
			loadedCfg, err := mgr.Load(configPath)

			// --- Assert
			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)
				}
				assert.Nil(t, loadedCfg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, loadedCfg)
				if tc.expectedFn != nil {
					tc.expectedFn(loadedCfg, tempDir) // Pass configDir for path assertions
				}
			}
		})
	}
}

func TestTomlManager_Save(t *testing.T) {
	mgr := config.NewTomlManager()
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "save-test.toml")

	// Create a domain.Config to save
	saveCfg := &domain.Config{
		Global: domain.GlobalConfig{
			CacheDir:  filepath.Join(configDir, ".cache"), // Use resolved paths for testing
			Namespace: "test-ns",
		},
		Inputs: map[string]domain.InputSource{
			"local1": {
				Type: "local",
				Details: domain.LocalInputSourceDetails{
					Path: filepath.Join(configDir, "local-rules"),
				},
			},
			"git1": {
				Type: "git",
				Details: domain.GitInputSourceDetails{
					Repository: "https://a.b/repo.git",
					Revision:   "dev",
					// SubDir is empty, should be omitted
				},
			},
		},
		Outputs: map[string]domain.OutputTarget{
			"cursor": {
				Target:  "cursor",
				Enabled: true, // Should be omitted on save
			},
			"vscode": {
				Target:  "vscode-copilot",
				Enabled: false, // Should be saved as false
			},
		},
	}

	// --- Act
	err := mgr.Save(configPath, saveCfg)
	require.NoError(t, err)

	// --- Assert: Read the saved file and verify its content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var loadedUserTomlCfg config.UserTomlConfig // Renamed variable type
	// Need to use toml.Unmarshal to read the file back
	err = toml.Unmarshal(data, &loadedUserTomlCfg)
	require.NoError(t, err)

	// Verify Global (pointers should exist and match)
	require.NotNil(t, loadedUserTomlCfg.Global)
	require.NotNil(t, loadedUserTomlCfg.Global.CacheDir)
	assert.Equal(t, saveCfg.Global.CacheDir, *loadedUserTomlCfg.Global.CacheDir)
	require.NotNil(t, loadedUserTomlCfg.Global.Namespace)
	assert.Equal(t, "test-ns", *loadedUserTomlCfg.Global.Namespace)

	// Verify Inputs
	require.Len(t, loadedUserTomlCfg.Inputs, 2)
	// Local input
	ucLocal, ok := loadedUserTomlCfg.Inputs["local1"]
	require.True(t, ok)
	assert.Equal(t, "local", ucLocal.Type)
	require.NotNil(t, ucLocal.Path)
	assert.Equal(t, saveCfg.Inputs["local1"].Details.(domain.LocalInputSourceDetails).Path, *ucLocal.Path)
	assert.Nil(t, ucLocal.Repository)
	// Git input
	ucGit, ok := loadedUserTomlCfg.Inputs["git1"]
	require.True(t, ok)
	assert.Equal(t, "git", ucGit.Type)
	require.NotNil(t, ucGit.Repository)
	assert.Equal(t, "https://a.b/repo.git", *ucGit.Repository)
	require.NotNil(t, ucGit.Revision) // Revision was not empty
	assert.Equal(t, "dev", *ucGit.Revision)
	assert.Nil(t, ucGit.SubDir) // SubDir was empty, should be nil
	assert.Nil(t, ucGit.Path)

	// Verify Outputs
	require.Len(t, loadedUserTomlCfg.Outputs, 2)
	// Cursor output (Enabled=true omitted)
	ucCursor, ok := loadedUserTomlCfg.Outputs["cursor"]
	require.True(t, ok)
	assert.Equal(t, "cursor", ucCursor.Target)
	assert.Nil(t, ucCursor.Enabled) // Enabled=true is omitted
	// VSCode output (Enabled=false saved)
	ucVscode, ok := loadedUserTomlCfg.Outputs["vscode"]
	require.True(t, ok)
	assert.Equal(t, "vscode-copilot", ucVscode.Target)
	require.NotNil(t, ucVscode.Enabled)
	assert.False(t, *ucVscode.Enabled)
}
