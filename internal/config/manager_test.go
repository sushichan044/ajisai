package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ai-rules-manager/internal/config"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

func TestLoad(t *testing.T) {
	t.Run("non-existent config returns fallback config", func(t *testing.T) {
		// テスト用の存在しないパスを指定
		nonExistentPath := filepath.Join(t.TempDir(), "non-existent.toml")

		cfg, err := config.Load(nonExistentPath)
		require.NoError(t, err) // 存在しないパスはエラーではなく、フォールバック設定を返す

		// フォールバック設定の内容を確認
		assert.Equal(t, "ai-rules-manager", cfg.Global.Namespace)
		assert.Empty(t, cfg.Inputs)
		assert.Empty(t, cfg.Outputs)
	})

	t.Run("unsupported extension returns error", func(t *testing.T) {
		// テスト用の.txtファイルを作成
		unsupportedPath := filepath.Join(t.TempDir(), "config.txt")
		err := os.WriteFile(unsupportedPath, []byte("test content"), 0644)
		require.NoError(t, err)

		_, err = config.Load(unsupportedPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported config file extension")
	})

	t.Run("valid toml file loads successfully", func(t *testing.T) {
		// テスト用の有効なTOMLファイルを作成
		validTomlPath := filepath.Join(t.TempDir(), "valid.toml")
		tomlContent := `
[global]
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

		cfg, err := config.Load(validTomlPath)
		require.NoError(t, err)

		// 読み込まれた設定の内容を確認
		assert.Equal(t, "test-namespace", cfg.Global.Namespace)
		assert.Contains(t, cfg.Inputs, "test")
		assert.Equal(t, "local", cfg.Inputs["test"].Type)
		assert.Contains(t, cfg.Outputs, "test")
		assert.Equal(t, "cursor", cfg.Outputs["test"].Target)
	})

	t.Run("invalid toml file returns error", func(t *testing.T) {
		// テスト用の不正なTOMLファイルを作成
		invalidTomlPath := filepath.Join(t.TempDir(), "invalid.toml")
		invalidContent := `
[global
namespace = "test" # 閉じ括弧がない
`
		err := os.WriteFile(invalidTomlPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		_, err = config.Load(invalidTomlPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal TOML")
	})
}

// MockConfigManager はテストのためのConfigManagerモック.
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

// テスト用モックマネージャーの検証.
func TestMockConfigManager(t *testing.T) {
	mockCfg := &domain.Config{
		Global: domain.GlobalConfig{
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

	// LoadFuncが期待通り動作するか検証
	loadedCfg, err := mock.Load("test-path")
	require.NoError(t, err)
	assert.Equal(t, mockCfg, loadedCfg)

	// SaveFuncが期待通り動作するか検証
	err = mock.Save("test-path", mockCfg)
	require.NoError(t, err)
}
