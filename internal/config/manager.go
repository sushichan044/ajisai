package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

type ConfigManager interface {
	Load(configPath string) (*domain.Config, error)
	Save(configPath string, cfg *domain.Config) error
}

var (
	fallbackConfig = domain.Config{
		Global: domain.GlobalConfig{
			Namespace: "ai-rules-manager",
			CacheDir:  "./.cache/ai-rules-manager",
		},
		Inputs:  make(map[string]domain.InputSource, 0),
		Outputs: make(map[string]domain.OutputTarget, 0),
	}
)

func CreateConfigManager() ConfigManager {
	return &ConfigManagerImpl{}
}

type ConfigManagerImpl struct{}

func (m *ConfigManagerImpl) Load(configPath string) (*domain.Config, error) {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	if _, err := os.Stat(resolvedPath); err != nil {
		// TODO: add warn log: "Failed to load config from %s, using fallback config", configPath
		return &fallbackConfig, nil
	}

	switch extension := filepath.Ext(resolvedPath); extension {
	case ".toml":
		return CreateTomlManager().Load(resolvedPath)
	default:
		return nil, fmt.Errorf("unsupported config file extension: %s", extension)
	}
}

func (m *ConfigManagerImpl) Save(configPath string, cfg *domain.Config) error {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	switch extension := filepath.Ext(resolvedPath); extension {
	case ".toml":
		return CreateTomlManager().Save(resolvedPath, cfg)
	default:
		return fmt.Errorf("unsupported config file extension: %s", extension)
	}
}
