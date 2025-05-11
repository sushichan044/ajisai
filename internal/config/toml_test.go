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
)

func TestTomlManager_Load_Success(t *testing.T) {
	tests := []struct {
		name           string
		tomlData       string
		expectedConfig *domain.Config
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
			expectedConfig: &domain.Config{
				Settings: domain.Settings{
					CacheDir:  "",
					Namespace: "",
				},
				Inputs: map[string]domain.InputSource{
					"local_rules": {
						Type: domain.InputSourceTypeLocal,
						Details: domain.LocalInputSourceDetails{
							Path: "./rules",
						},
					},
				},
				Outputs: map[string]domain.OutputTarget{
					"cursor": {
						Target:  domain.OutputTargetTypeCursor,
						Enabled: true,
					},
				},
			},
		},
		{
			name: "full config with settings and git",
			tomlData: `
[settings]
cacheDir = "~/.cache/ai-rules"
namespace = "my-proj"
experimental = true

[inputs.remote_rules]
type = "git"
repository = "https://example.com/repo.git"
revision = "main"
directory = "presets"

[outputs.github_copilot]
target = "github-copilot"
enabled = false
`,
			expectedConfig: &domain.Config{
				Settings: domain.Settings{
					CacheDir:     "~/.cache/ai-rules",
					Namespace:    "my-proj",
					Experimental: true,
				},
				Inputs: map[string]domain.InputSource{
					"remote_rules": {
						Type: domain.InputSourceTypeGit,
						Details: domain.GitInputSourceDetails{
							Repository: "https://example.com/repo.git",
							Revision:   "main",
							Directory:  "presets",
						},
					},
				},
				Outputs: map[string]domain.OutputTarget{
					"github_copilot": {
						Target:  domain.OutputTargetTypeGitHubCopilot,
						Enabled: false,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mgr := config.NewTomlLoader()

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

			assert.Equal(t, tc.expectedConfig.Settings, loadedCfg.Settings)
			assert.Equal(t, tc.expectedConfig.Inputs, loadedCfg.Inputs)
			assert.Equal(t, tc.expectedConfig.Outputs, loadedCfg.Outputs)
		})
	}
}

func TestTomlManager_Load_FileNotFound(t *testing.T) {
	mgr := config.NewTomlLoader()
	configPath := filepath.Join(t.TempDir(), "non_existent_file.toml")

	// --- Act
	_, err := mgr.Load(configPath)

	// --- Assert
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to read config file")
	assert.ErrorIs(t, err, os.ErrNotExist) // Check underlying error
}

func TestTomlManager_Load_InvalidToml(t *testing.T) {
	mgr := config.NewTomlLoader()
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
}

func TestTomlManager_Save(t *testing.T) {
	mgr := config.NewTomlLoader()
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "save-test.toml")

	// Create a domain.Config to save
	saveCfg := &domain.Config{
		Settings: domain.Settings{
			CacheDir:     filepath.Join(configDir, ".cache"), // Use resolved paths for testing
			Namespace:    "test-ns",
			Experimental: true,
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
			"github_copilot": {
				Target:  domain.OutputTargetTypeGitHubCopilot,
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

	var loadedUserTomlCfg config.UserTomlConfig
	// Need to use toml.Unmarshal to read the file back
	err = toml.Unmarshal(data, &loadedUserTomlCfg)
	require.NoError(t, err)

	expectedUserTomlCfg := config.NewTomlLoader().ToFormat(saveCfg)

	assert.Equal(t, expectedUserTomlCfg.Settings, loadedUserTomlCfg.Settings)
	assert.Equal(t, expectedUserTomlCfg.Inputs, loadedUserTomlCfg.Inputs)
	assert.Equal(t, expectedUserTomlCfg.Outputs, loadedUserTomlCfg.Outputs)
}
