package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/aisync/internal/config"
	"github.com/sushichan044/aisync/internal/domain"
	"github.com/sushichan044/aisync/internal/utils"
)

// Helper function from Validation test.
func absPath(t *testing.T, relative string, baseDir string) string {
	abs, err := filepath.Abs(filepath.Join(baseDir, relative))
	require.NoError(t, err)
	return abs
}

func TestTomlManager_Load_Success(t *testing.T) {
	homeDir, _ := os.UserHomeDir() // For ~ expansion test
	require.NotEmpty(t, homeDir)

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
enabled = true
`,
			expectedFn: func(cfg *domain.Config, configDir string) {
				// default namespace
				assert.Equal(t, "aisync", cfg.Global.Namespace)
				// default cache dir
				assert.Equal(t, absPath(t, "./.cache/aisync", configDir), cfg.Global.CacheDir)

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
directory = "presets"

[outputs.vscode]
target = "vscode-copilot"
enabled = false
`,
			expectedFn: func(cfg *domain.Config, _ string) {
				assert.Equal(t, "my-proj", cfg.Global.Namespace)

				resolvedPath, err := utils.ResolveAbsPath("~/.cache/ai-rules")
				require.NoError(t, err)
				assert.Equal(t, resolvedPath, cfg.Global.CacheDir)

				require.Len(t, cfg.Inputs, 1)
				input, ok := cfg.Inputs["remote_rules"]
				require.True(t, ok)
				assert.Equal(t, "git", input.Type)
				details, ok := input.Details.(domain.GitInputSourceDetails)
				require.True(t, ok, "Details should be GitInputSourceDetails")
				assert.Equal(t, "https://example.com/repo.git", details.Repository)
				assert.Equal(t, "main", details.Revision)
				assert.Equal(t, "presets", details.Directory)

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
			mgr := config.CreateTomlManager()

			// Create a temporary file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "aisync.toml")
			err := os.WriteFile(configPath, []byte(tc.tomlData), 0644)
			require.NoError(t, err)

			// --- Act
			loadedCfg, err := mgr.Load(configPath)

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
	mgr := config.CreateTomlManager()
	configPath := filepath.Join(t.TempDir(), "non_existent_file.toml")

	// --- Act
	_, err := mgr.Load(configPath)

	// --- Assert
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to read config file")
	assert.ErrorIs(t, err, os.ErrNotExist) // Check underlying error
}

func TestTomlManager_Load_InvalidToml(t *testing.T) {
	mgr := config.CreateTomlManager()
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

func TestTomlManager_Save(t *testing.T) {
	mgr := config.CreateTomlManager()
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
				Enabled: true,
			},
			"vscode": {
				Target:  "vscode-copilot",
				Enabled: false,
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

	// Verify Global
	assert.Equal(t, saveCfg.Global.CacheDir, loadedUserTomlCfg.Global.CacheDir)
	assert.Equal(t, "test-ns", loadedUserTomlCfg.Global.Namespace)

	// Verify Inputs
	require.Len(t, loadedUserTomlCfg.Inputs, 2)
	// Local input
	ucLocal, ok := loadedUserTomlCfg.Inputs["local1"]
	require.True(t, ok)
	assert.Equal(t, "local", ucLocal.Type)
	assert.Equal(t, saveCfg.Inputs["local1"].Details.(domain.LocalInputSourceDetails).Path, ucLocal.Path)
	assert.Empty(t, ucLocal.Repository)

	// Git input
	ucGit, ok := loadedUserTomlCfg.Inputs["git1"]
	require.True(t, ok)
	assert.Equal(t, "git", ucGit.Type)
	assert.Equal(t, "https://a.b/repo.git", ucGit.Repository)
	assert.Equal(t, "dev", ucGit.Revision)
	assert.Empty(t, ucGit.Directory)

	// Verify Outputs
	require.Len(t, loadedUserTomlCfg.Outputs, 2)
	// Cursor output (Enabled=true omitted)
	ucCursor, ok := loadedUserTomlCfg.Outputs["cursor"]
	require.True(t, ok)
	assert.Equal(t, "cursor", ucCursor.Target)
	assert.True(t, ucCursor.Enabled)

	// VSCode output (Enabled=false saved)
	ucVscode, ok := loadedUserTomlCfg.Outputs["vscode"]
	require.True(t, ok)
	assert.Equal(t, "vscode-copilot", ucVscode.Target)
	assert.False(t, ucVscode.Enabled)
}
