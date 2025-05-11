package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/sushichan044/aisync/internal/domain"
)

type TomlLoader struct{}

func NewTomlLoader() formatLoader[UserTomlConfig] {
	return &TomlLoader{}
}

type (
	UserTomlConfig struct {
		Global  UserTomlGlobalConfig            `toml:"global,omitempty"`
		Inputs  map[string]UserTomlInputSource  `toml:"inputs,omitempty"`
		Outputs map[string]UserTomlOutputTarget `toml:"outputs,omitempty"`
	}

	UserTomlGlobalConfig struct {
		CacheDir  string `toml:"cacheDir,omitempty"`
		Namespace string `toml:"namespace,omitempty"`
	}

	UserTomlInputSource struct {
		Type       domain.InputSourceType `toml:"type"`                 // Required
		Path       string                 `toml:"path,omitempty"`       // Used if type=local
		Repository string                 `toml:"repository,omitempty"` // Used if type=git
		Revision   string                 `toml:"revision,omitempty"`   // Used if type=git (Optional ref/branch/tag/commit)
		Directory  string                 `toml:"directory,omitempty"`  // Used if type=git (Optional)
	}

	UserTomlOutputTarget struct {
		Target  domain.OutputTargetType `toml:"target"`
		Enabled bool                    `toml:"enabled"`
	}
)

// Load a config from given path. Returns a fallback config if the path is invalid.
func (loader *TomlLoader) Load(configPath string) (*domain.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var userTomlCfg UserTomlConfig
	err = toml.Unmarshal(data, &userTomlCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML config from %s: %w", configPath, err)
	}

	return loader.FromFormat(userTomlCfg), nil
}

func (loader *TomlLoader) Save(configPath string, cfg *domain.Config) error {
	tomlCfg := loader.ToFormat(cfg)

	// Create parent directories if they don't exist
	configDir := filepath.Dir(configPath)
	if ensureDirErr := os.MkdirAll(configDir, 0750); ensureDirErr != nil {
		return fmt.Errorf("failed to create directory %s: %w", configDir, ensureDirErr)
	}

	// Write to a temporary file first for atomicity
	tempFile, err := os.CreateTemp(configDir, ".aisync.tmp-")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file on error or success
	defer tempFile.Close()           // Ensure file is closed

	// Marshal to TOML and write to temp file
	encoder := toml.NewEncoder(tempFile)
	if encodeErr := encoder.Encode(tomlCfg); encodeErr != nil {
		return fmt.Errorf("failed to encode config to TOML: %w", encodeErr)
	}

	// Rename the temporary file to the final config path
	if renameErr := os.Rename(tempFile.Name(), configPath); renameErr != nil {
		return renameErr
	}

	return nil
}

func (loader *TomlLoader) ToFormat(cfg *domain.Config) UserTomlConfig {
	inputs := make(map[string]UserTomlInputSource)
	for key, input := range cfg.Inputs {
		switch inputDetails := input.Details.(type) {
		case domain.LocalInputSourceDetails:
			inputs[key] = UserTomlInputSource{
				Type: domain.InputSourceTypeLocal,
				Path: inputDetails.Path,
			}
		case domain.GitInputSourceDetails:
			inputs[key] = UserTomlInputSource{
				Type:       domain.InputSourceTypeGit,
				Repository: inputDetails.Repository,
				Revision:   inputDetails.Revision,
				Directory:  inputDetails.Directory,
			}
		}
	}

	outputs := make(map[string]UserTomlOutputTarget)
	for key, output := range cfg.Outputs {
		outputs[key] = UserTomlOutputTarget{
			Target:  output.Target,
			Enabled: output.Enabled,
		}
	}
	return UserTomlConfig{
		Global: UserTomlGlobalConfig{
			CacheDir:  cfg.Global.CacheDir,
			Namespace: cfg.Global.Namespace,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func (loader *TomlLoader) FromFormat(userTomlCfg UserTomlConfig) *domain.Config {
	inputs := make(map[string]domain.InputSource)
	for key, input := range userTomlCfg.Inputs {
		switch input.Type {
		case domain.InputSourceTypeLocal:
			inputs[key] = domain.InputSource{
				Type: domain.InputSourceTypeLocal,
				Details: domain.LocalInputSourceDetails{
					Path: input.Path,
				},
			}
		case domain.InputSourceTypeGit:
			inputs[key] = domain.InputSource{
				Type: domain.InputSourceTypeGit,
				Details: domain.GitInputSourceDetails{
					Repository: input.Repository,
					Revision:   input.Revision,
					Directory:  input.Directory,
				},
			}
		}
	}

	outputs := make(map[string]domain.OutputTarget)
	for key, output := range userTomlCfg.Outputs {
		outputs[key] = domain.OutputTarget{
			Target:  output.Target,
			Enabled: output.Enabled,
		}
	}

	return &domain.Config{
		Global: domain.GlobalConfig{
			Namespace: userTomlCfg.Global.Namespace,
			CacheDir:  userTomlCfg.Global.CacheDir,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}
}
