package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
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

func resolveConfigPath(configPath string) (string, error) {
	if filepath.IsAbs(configPath) {
		return configPath, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	return filepath.Join(wd, configPath), nil
}

// Load a config from given path. Returns a fallback config if the path is invalid.
func Load(configPath string) (*domain.Config, error) {
	resolvedPath, err := resolveConfigPath(configPath)
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
