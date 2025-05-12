package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

// Format-specific config loaders must implement this interface.
type formatLoader[TFormat any] interface {
	Load(configPath string) (*domain.Config, error)

	Save(configPath string, cfg *domain.Config) error

	ToFormat(cfg *domain.Config) TFormat
	FromFormat(format TFormat) *domain.Config
}

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Load(configPath string) (*domain.Config, error) {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	if _, statErr := os.Stat(resolvedPath); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			// If the config file does not exist, return a default config
			return m.ApplyDefaults(nil)
		}
		return nil, fmt.Errorf("failed to stat config file %s: %w", resolvedPath, statErr)
	}

	var loadCfg *domain.Config
	switch extension := filepath.Ext(resolvedPath); extension {
	case ".toml":
		tomlCfg, tomlErr := NewTomlLoader().Load(resolvedPath)
		if tomlErr != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", resolvedPath, tomlErr)
		}
		loadCfg = tomlCfg
	default:
		return nil, fmt.Errorf("unsupported config file extension: %s", extension)
	}

	return m.ApplyDefaults(loadCfg)
}

func (m *Manager) Save(configPath string, cfg *domain.Config) error {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	switch extension := filepath.Ext(resolvedPath); extension {
	case ".toml":
		return NewTomlLoader().Save(resolvedPath, cfg)
	default:
		return fmt.Errorf("unsupported config file extension: %s", extension)
	}
}

func (m *Manager) ApplyDefaults(cfg *domain.Config) (*domain.Config, error) {
	defaultCfg := m.GetDefaultConfig()

	if cfg == nil {
		return defaultCfg, nil
	}

	if cfg.Settings == (domain.Settings{}) {
		cfg.Settings = defaultCfg.Settings
	} else {
		if cfg.Settings.Namespace == "" {
			cfg.Settings.Namespace = defaultCfg.Settings.Namespace
		}
		if cfg.Settings.CacheDir == "" {
			cfg.Settings.CacheDir = defaultCfg.Settings.CacheDir
		}
		// Experimental is bool, so it defaults to false if not set. We don't need to set it.
	}

	if cfg.Inputs == nil {
		cfg.Inputs = defaultCfg.Inputs
	}
	if cfg.Outputs == nil {
		cfg.Outputs = defaultCfg.Outputs
	}
	return cfg, nil
}

func (m *Manager) GetDefaultConfig() *domain.Config {
	return &domain.Config{
		Settings: domain.Settings{
			Namespace:    "ajisai",
			CacheDir:     "./.cache/ajisai",
			Experimental: false,
		},
		Inputs:  make(map[string]domain.InputSource),
		Outputs: make(map[string]domain.OutputTarget),
	}
}
