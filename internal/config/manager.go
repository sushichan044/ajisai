package config

import (
	"fmt"
	"os"
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

type Manager struct {
	candidateConfigPaths []string
}

// NewManager creates a new Manager that reads the config files from the given paths.
func NewManager(candidateAbsPaths ...string) (*Manager, error) {
	supportedPaths := make([]string, 0, len(candidateAbsPaths))
	for _, configPath := range candidateAbsPaths {
		if !isSupportedConfigFilePath(configPath) {
			return nil, &UnsupportedConfigFileError{
				Path: configPath,
			}
		}

		supportedPaths = append(supportedPaths, configPath)
	}

	return &Manager{
		candidateConfigPaths: supportedPaths,
	}, nil
}

// NewDefaultManagerInDir creates a new Manager that reads the default config files in the given directory.
func NewDefaultManagerInDir(dir string) (*Manager, error) {
	absDir, err := utils.ResolveAbsPath(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	defaultYml := filepath.Join(absDir, defaultConfigFileYml)
	defaultYaml := filepath.Join(absDir, defaultConfigFileYaml)

	return NewManager(defaultYaml, defaultYml)
}

// Load loads the config from the config paths. It returns the first valid config file.
func (m *Manager) Load() (*Config, error) {
	targetPath, err := m.getFileToRead()
	if err != nil {
		return nil, err
	}

	loadedCfg, err := newYAMLLoader().Load(targetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file %s: %w", targetPath, err)
	}

	return m.ApplyDefaults(loadedCfg)
}

func (m *Manager) Save(cfg *Config) error {
	targetPath, err := m.getFileToWrite()
	if err != nil {
		return err
	}

	return newYAMLLoader().Save(targetPath, cfg)
}

// getFileToRead returns a readable config file path.
// It returns the first existing file.
//
// If no existing file is found, it returns a NoFileToReadError.
func (m *Manager) getFileToRead() (string, error) {
	for _, configPath := range m.candidateConfigPaths {
		if _, statErr := os.Stat(configPath); statErr == nil {
			return configPath, nil
		}
	}

	return "", &NoFileToReadError{CandidateConfigPaths: m.candidateConfigPaths}
}

// getFileToWrite returns a writable config file path.
// It returns the first existing file path or falls back to the first candidate path.
func (m *Manager) getFileToWrite() (string, error) {
	if len(m.candidateConfigPaths) == 0 {
		return "", &NoFileToWriteError{}
	}

	for _, configPath := range m.candidateConfigPaths {
		if _, statErr := os.Stat(configPath); statErr == nil {
			return configPath, nil
		}
	}

	return m.candidateConfigPaths[0], nil
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
