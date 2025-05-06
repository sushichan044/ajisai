package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

func CreateTomlManager() ConfigManager {
	return &TomlManager{}
}

type TomlManager struct{}

// Load a config from given path. Returns a fallback config if the path is invalid.
func (m *TomlManager) Load(configPath string) (*domain.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var userTomlCfg UserTomlConfig
	err = toml.Unmarshal(data, &userTomlCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML config from %s: %w", configPath, err)
	}

	internalCfg, err := transformAndValidate(userTomlCfg, configPath)
	if err != nil {
		// Wrap the transformation/validation error for context
		return nil, fmt.Errorf("error processing config from %s: %w", configPath, err)
	}

	return internalCfg, nil
}

func (m *TomlManager) Save(configPath string, cfg *domain.Config) error {
	userTomlCfg, err := domainConfigToUserTomlConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to convert domain config to user config for saving: %w", err)
	}

	// Create parent directories if they don't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configDir, err)
	}

	// Write to a temporary file first for atomicity
	tempFile, err := os.CreateTemp(configDir, ".ai-rules.tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temporary config file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up temp file on error or success
	defer tempFile.Close()           // Ensure file is closed

	// Marshal to TOML and write to temp file
	encoder := toml.NewEncoder(tempFile)
	if err := encoder.Encode(userTomlCfg); err != nil {
		tempFile.Close() // Close before removing
		return fmt.Errorf("failed to encode config to TOML: %w", err)
	}

	// Close the file explicitly before renaming
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary config file: %w", err)
	}

	// Rename the temporary file to the final config path
	if err := os.Rename(tempFile.Name(), configPath); err != nil {
		return fmt.Errorf("failed to rename temporary config file to %s: %w", configPath, err)
	}

	return nil
}

func transformAndValidate(userTomlCfg UserTomlConfig, configFilePath string) (*domain.Config, error) {
	configDir := filepath.Dir(configFilePath)

	globalCfg, err := processGlobalConfig(userTomlCfg.Global, configDir)
	if err != nil {
		return nil, fmt.Errorf("error processing global config: %w", err)
	}

	inputsMap, err := processInputs(userTomlCfg.Inputs, configDir)
	if err != nil {
		return nil, err
	}

	outputsMap, err := processOutputs(userTomlCfg.Outputs)
	if err != nil {
		return nil, err
	}

	cfg := &domain.Config{
		Global:  globalCfg,
		Inputs:  inputsMap,
		Outputs: outputsMap,
	}

	return cfg, nil
}

// processGlobalConfig sets default values for GlobalConfig and resolves paths.
// Returns the processed GlobalConfig.
func processGlobalConfig(userTomlGlobal *UserTomlGlobalConfig, configDir string) (domain.GlobalConfig, error) {
	// Defaults
	defaultNamespace := "default"
	defaultCacheDir := "./.cache/ai-rules-manager" // Relative to config file

	namespace := defaultNamespace
	cacheDir := defaultCacheDir

	if userTomlGlobal != nil {
		if userTomlGlobal.Namespace != nil && *userTomlGlobal.Namespace != "" {
			namespace = *userTomlGlobal.Namespace
		}
		if userTomlGlobal.CacheDir != nil && *userTomlGlobal.CacheDir != "" {
			cacheDir = *userTomlGlobal.CacheDir
		}
	}

	var resolvedCacheDir string
	var err error

	if filepath.IsAbs(cacheDir) || strings.HasPrefix(cacheDir, "~") {
		resolvedCacheDir, err = utils.ResolveAbsPath(cacheDir)
		if err != nil {
			return domain.GlobalConfig{}, fmt.Errorf("failed to resolve cache directory path: %w", err)
		}
	} else {
		resolvedCacheDir, err = utils.ResolveAbsPath(filepath.Join(configDir, cacheDir))
		if err != nil {
			return domain.GlobalConfig{}, fmt.Errorf("failed to resolve cache directory path: %w", err)
		}
	}

	return domain.GlobalConfig{
		Namespace: namespace,
		CacheDir:  filepath.Clean(resolvedCacheDir),
	}, nil
}

// processInputs transforms and validates the input sources.
// Returns the processed map of domain.InputSource or an error.
func processInputs(
	userTomlInputs map[string]UserTomlInputSource,
	configDir string,
) (map[string]domain.InputSource, error) {
	if userTomlInputs == nil {
		return make(map[string]domain.InputSource), nil // Return empty map if none defined
	}

	processedInputs := make(map[string]domain.InputSource, len(userTomlInputs))

	for key, userInput := range userTomlInputs {
		if userInput.Type == "" {
			return nil, fmt.Errorf("input source '%s': missing required 'type' field", key)
		}

		var details domain.InputSourceDetails
		var err error

		switch userInput.Type {
		case "local":
			details, err = validateLocalInput(userInput, configDir)
		case "git":
			details, err = validateGitInput(userInput)
		default:
			return nil, fmt.Errorf("input source '%s': unsupported type '%s'", key, userInput.Type)
		}

		if err != nil {
			// Wrap error with context
			return nil, fmt.Errorf("input source '%s': %w", key, err)
		}

		processedInputs[key] = domain.InputSource{
			Type:    userInput.Type,
			Details: details,
		}
	}
	return processedInputs, nil
}

// validateLocalInput validates a local input source and resolves its path.
func validateLocalInput(userInput UserTomlInputSource, configDir string) (domain.LocalInputSourceDetails, error) {
	if userInput.Path == nil || *userInput.Path == "" {
		return domain.LocalInputSourceDetails{}, errors.New("type 'local' requires 'path' field")
	}
	if userInput.Repository != nil || userInput.Revision != nil || userInput.SubDir != nil {
		return domain.LocalInputSourceDetails{}, errors.New(
			"type 'local' does not support 'repository', 'revision', or 'subDir' fields",
		)
	}

	localPath := *userInput.Path

	var resolvedPath string
	var err error

	if filepath.IsAbs(localPath) || strings.HasPrefix(localPath, "~") {
		resolvedPath, err = utils.ResolveAbsPath(localPath)
	} else {
		resolvedPath, err = utils.ResolveAbsPath(filepath.Join(configDir, localPath))
	}

	if err != nil {
		return domain.LocalInputSourceDetails{}, fmt.Errorf("failed to resolve path: %w", err)
	}

	return domain.LocalInputSourceDetails{Path: filepath.Clean(resolvedPath)}, nil
}

// validateGitInput validates a git input source.
func validateGitInput(userInput UserTomlInputSource) (domain.GitInputSourceDetails, error) {
	if userInput.Repository == nil || *userInput.Repository == "" {
		return domain.GitInputSourceDetails{}, errors.New("type 'git' requires 'repository' field")
	}
	if userInput.Path != nil {
		return domain.GitInputSourceDetails{}, errors.New("type 'git' does not support 'path' field")
	}

	details := domain.GitInputSourceDetails{
		Repository: *userInput.Repository,
	}
	if userInput.Revision != nil {
		details.Revision = *userInput.Revision
	}
	if userInput.SubDir != nil {
		details.SubDir = *userInput.SubDir
	}
	return details, nil
}

// processOutputs transforms and validates the output targets.
// Returns the processed map of domain.OutputTarget or an error.
func processOutputs(userTomlOutputs map[string]UserTomlOutputTarget) (map[string]domain.OutputTarget, error) {
	if userTomlOutputs == nil {
		return make(map[string]domain.OutputTarget), nil // Return empty map if none defined
	}

	processedOutputs := make(map[string]domain.OutputTarget, len(userTomlOutputs))

	for key, userOutput := range userTomlOutputs {
		if userOutput.Target == "" {
			return nil, fmt.Errorf("output target '%s': missing required 'target' field", key)
		}

		// TODO: Validate userOutput.Target against known adapter types?

		enabled := true // Default to true if omitted
		if userOutput.Enabled != nil {
			enabled = *userOutput.Enabled
		}

		processedOutputs[key] = domain.OutputTarget{
			Target:  userOutput.Target,
			Enabled: enabled,
			// Details: // Add when needed
		}
	}
	return processedOutputs, nil
}

// domainConfigToUserTomlConfig converts the internal domain.Config back to the user-facing
// UserTomlConfig structure, suitable for saving to TOML.
func domainConfigToUserTomlConfig(cfg *domain.Config) (*UserTomlConfig, error) {
	userTomlCfg := &UserTomlConfig{
		Global: &UserTomlGlobalConfig{
			// Pointers are needed for user config
			CacheDir:  &cfg.Global.CacheDir,
			Namespace: &cfg.Global.Namespace,
		},
		Inputs:  make(map[string]UserTomlInputSource),
		Outputs: make(map[string]UserTomlOutputTarget),
	}

	// TODO: Consider if we should omit default values on save?
	// Current approach saves resolved values.
	if userTomlCfg.Global.CacheDir != nil && *userTomlCfg.Global.CacheDir == "" {
		userTomlCfg.Global.CacheDir = nil
	}
	if userTomlCfg.Global.Namespace != nil && *userTomlCfg.Global.Namespace == "" {
		userTomlCfg.Global.Namespace = nil
	}
	if userTomlCfg.Global.CacheDir == nil && userTomlCfg.Global.Namespace == nil {
		userTomlCfg.Global = nil // Omit global section if both are empty/default? Decide this.
	}

	for key, input := range cfg.Inputs {
		ucInput := UserTomlInputSource{
			Type: input.Type,
		}
		switch d := input.Details.(type) {
		case domain.LocalInputSourceDetails:
			ucInput.Path = &d.Path
		case domain.GitInputSourceDetails:
			ucInput.Repository = &d.Repository
			if d.Revision != "" { // Save revision only if not empty
				ucInput.Revision = &d.Revision
			}
			if d.SubDir != "" { // Save subDir only if not empty
				ucInput.SubDir = &d.SubDir
			}
		default:
			return nil, fmt.Errorf("input source '%s': unknown details type %T during conversion", key, input.Details)
		}
		userTomlCfg.Inputs[key] = ucInput
	}

	for key, output := range cfg.Outputs {
		ucOutput := UserTomlOutputTarget{
			Target: output.Target,
		}
		// Save enabled flag only if it's false (since true is the default)
		if !output.Enabled {
			enabledFalse := false
			ucOutput.Enabled = &enabledFalse
		}
		userTomlCfg.Outputs[key] = ucOutput
	}

	return userTomlCfg, nil
}
