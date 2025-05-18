package config

import (
	"fmt"
	"path/filepath"

	"github.com/sushichan044/ajisai/utils"
)

// Format-specific config loaders must implement this interface.
type formatLoader[T any] interface {
	Load(configPath string) (*Config, error)
	Save(configPath string, cfg *Config) error

	toFormat(cfg *Config) (T, error)
	fromFormat(cfg T) (*Config, error)
}

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Load(configPath string) (*Config, error) {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	var loadCfg *Config
	switch extension := filepath.Ext(resolvedPath); extension {
	case ".json":
		loadCfg, err = newJSONLoader().Load(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", resolvedPath, err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file extension: %s", extension)
	}

	return m.ApplyDefaults(loadCfg)
}

func (m *Manager) Save(configPath string, cfg *Config) error {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	switch extension := filepath.Ext(resolvedPath); extension {
	case ".json":
		return newJSONLoader().Save(resolvedPath, cfg)
	default:
		return fmt.Errorf("unsupported config file extension: %s", extension)
	}
}

func (m *Manager) ApplyDefaults(cfg *Config) (*Config, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.Package = applyDefaultsToPackage(cfg.Package)
	cfg.Settings = applyDefaultsToSettings(cfg.Settings)
	cfg.Workspace = applyDefaultsToWorkspace(cfg.Workspace)

	return cfg, nil
}

func (m *Manager) GetDefaultConfig() *Config {
	return &Config{
		Settings:  applyDefaultsToSettings(nil),
		Package:   applyDefaultsToPackage(nil),
		Workspace: applyDefaultsToWorkspace(nil),
	}
}
