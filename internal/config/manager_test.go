package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/aisync/internal/config"
	"github.com/sushichan044/aisync/internal/domain"
)

func TestLoad(t *testing.T) {
	t.Run("non-existent config returns fallback config", func(t *testing.T) {
		nonExistentPath := filepath.Join(t.TempDir(), "non-existent.toml")

		cfg, err := config.NewManager().Load(nonExistentPath)

		// Non-existent path does not return an error, but returns a fallback config
		require.NoError(t, err)

		assert.Equal(t, "aisync", cfg.Settings.Namespace)
		assert.Empty(t, cfg.Inputs)
		assert.Empty(t, cfg.Outputs)
	})

	t.Run("unsupported extension returns error", func(t *testing.T) {
		unsupportedPath := filepath.Join(t.TempDir(), "config.txt")
		err := os.WriteFile(unsupportedPath, []byte("test content"), 0644)
		require.NoError(t, err)

		_, err = config.NewManager().Load(unsupportedPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported config file extension")
	})

	t.Run("valid toml file loads successfully", func(t *testing.T) {
		validTomlPath := filepath.Join(t.TempDir(), "valid.toml")
		tomlContent := `
[settings]
namespace = "test-namespace"
cacheDir = "./test-cache"

[inputs.test]
type = "local"
path = "./test-path"

[outputs.test]
target = "cursor"
`
		err := os.WriteFile(validTomlPath, []byte(tomlContent), 0644)
		require.NoError(t, err)

		cfg, err := config.NewManager().Load(validTomlPath)
		require.NoError(t, err)

		assert.Equal(t, "test-namespace", cfg.Settings.Namespace)
		assert.Contains(t, cfg.Inputs, "test")
		assert.Equal(t, domain.InputSourceTypeLocal, cfg.Inputs["test"].Type)
		assert.Contains(t, cfg.Outputs, "test")
		assert.Equal(t, domain.OutputTargetTypeCursor, cfg.Outputs["test"].Target)
	})

	t.Run("invalid toml file returns error", func(t *testing.T) {
		invalidTomlPath := filepath.Join(t.TempDir(), "invalid.toml")
		invalidContent := `
[settings
namespace = "test" # No closing bracket! Syntax error!
`
		err := os.WriteFile(invalidTomlPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		_, err = config.NewManager().Load(invalidTomlPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal TOML")
	})
}

type MockConfigManager struct {
	LoadFunc func(configPath string) (*domain.Config, error)
	SaveFunc func(configPath string, cfg *domain.Config) error
}

func (m *MockConfigManager) Load(configPath string) (*domain.Config, error) {
	return m.LoadFunc(configPath)
}

func (m *MockConfigManager) Save(configPath string, cfg *domain.Config) error {
	return m.SaveFunc(configPath, cfg)
}

func TestMockConfigManager(t *testing.T) {
	mockCfg := &domain.Config{
		Settings: domain.Settings{
			Namespace: "mock-namespace",
			CacheDir:  "./mock-cache",
		},
	}

	mock := &MockConfigManager{
		LoadFunc: func(configPath string) (*domain.Config, error) {
			assert.Equal(t, "test-path", configPath)
			return mockCfg, nil
		},
		SaveFunc: func(configPath string, cfg *domain.Config) error {
			assert.Equal(t, "test-path", configPath)
			assert.Equal(t, mockCfg, cfg)
			return nil
		},
	}

	// Verify that LoadFunc behaves as expected
	loadedCfg, err := mock.Load("test-path")
	require.NoError(t, err)
	assert.Equal(t, mockCfg, loadedCfg)

	// Verify that SaveFunc behaves as expected
	err = mock.Save("test-path", mockCfg)
	require.NoError(t, err)
}
