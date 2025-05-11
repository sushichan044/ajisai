package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sushichan044/aisync/internal/domain"
	"github.com/sushichan044/aisync/internal/utils"
)

func CreateConfigManager() domain.ConfigManager {
	return &configManagerImpl{}
}

type configManagerImpl struct{}

func (m *configManagerImpl) Load(configPath string) (*domain.Config, error) {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	if _, statErr := os.Stat(resolvedPath); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			// TODO: add warn log: "Failed to load config from %s, using fallback config", configPath
			// Return a fallback config.
			return &domain.Config{
				Global: domain.GlobalConfig{
					Namespace: "aisync",
					CacheDir:  "./.cache/aisync",
				},
				Inputs:  make(map[string]domain.InputSource, 0),
				Outputs: make(map[string]domain.OutputTarget, 0),
			}, nil
		}

		return nil, statErr
	}

	switch extension := filepath.Ext(resolvedPath); extension {
	case ".toml":
		return CreateTomlManager().Load(resolvedPath)
	default:
		return nil, fmt.Errorf("unsupported config file extension: %s", extension)
	}
}

func (m *configManagerImpl) Save(configPath string, cfg *domain.Config) error {
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
