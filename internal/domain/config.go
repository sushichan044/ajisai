package domain

import (
	"fmt"
	"path/filepath"

	"github.com/sushichan044/ajisai/utils"
)

const (
	PresetSourceTypeLocal PresetSourceType = "local" // Local file system input
	PresetSourceTypeGit   PresetSourceType = "git"   // Git repository input

	SupportedAgentTypeCursor        SupportedAgentType = "cursor"         // Cursor output target
	SupportedAgentTypeGitHubCopilot SupportedAgentType = "github-copilot" // GitHub Copilot output target
	SupportedAgentTypeWindsurf      SupportedAgentType = "windsurf"       // WindSurf output target
)

type (
	PresetSourceType   string
	SupportedAgentType string

	// Config represents the fully resolved and validated application configuration.
	Config struct {
		Settings                         // Resolved top-level settings
		Inputs   map[string]InputSource  // Key is the input source identifier
		Outputs  map[string]OutputTarget // Key is the output target identifier
	}

	// Settings holds application-wide settings with defaults applied.
	Settings struct {
		CacheDir     string // Resolved cache directory path
		Namespace    string // Resolved namespace
		Experimental bool   // Experimental features enabled
	}

	// InputSource defines a configured source for presets.
	InputSource struct {
		Type    PresetSourceType   // Type identifier (e.g., "local", "git")
		Details InputSourceDetails // Type-specific configuration details
	}

	// InputSourceDetails is an interface for type-specific input source configurations.
	InputSourceDetails interface {
		isInputSourceDetails()
	}

	LocalInputSourceDetails struct {
		Path string // Path to the local directory
	}

	// GitInputSourceDetails holds configuration specific to Git repository inputs.
	GitInputSourceDetails struct {
		Repository string // URL of the Git repository
		Revision   string // Optional branch, tag, or commit SHA (defaults resolved by Fetcher)
		Directory  string // Optional subdirectory within the repo
	}

	// OutputTarget defines a configured destination for the processed presets.
	OutputTarget struct {
		Target  SupportedAgentType // Type of output target (e.g., "cursor", "github-copilot")
		Enabled bool
	}
)

func (d LocalInputSourceDetails) isInputSourceDetails() {}

func (d GitInputSourceDetails) isInputSourceDetails() {}

// GetInputSourceDetails safely performs a type assertion on InputSource.Details.
func GetInputSourceDetails[T InputSourceDetails](is InputSource) (T, bool) {
	details, ok := is.Details.(T)
	return details, ok
}

func (c *Config) GetPresetRootInCache(presetName string) (string, error) {
	cacheDir, err := utils.ResolveAbsPath(c.Settings.CacheDir)
	if err != nil {
		return "", err
	}

	inputConfig, isConfigured := c.Inputs[presetName]
	if !isConfigured {
		return "", fmt.Errorf("preset %s not found", presetName)
	}

	switch inputConfig.Type {
	case PresetSourceTypeLocal:
		if _, ok := GetInputSourceDetails[LocalInputSourceDetails](inputConfig); ok {
			return filepath.Join(cacheDir, presetName), nil
		}
		return "", fmt.Errorf("invalid input source type: %s", inputConfig.Type)
	case PresetSourceTypeGit:
		if gitDetails, ok := GetInputSourceDetails[GitInputSourceDetails](inputConfig); ok {
			if gitDetails.Directory != "" {
				return filepath.Join(cacheDir, presetName, gitDetails.Directory), nil
			}
			return filepath.Join(cacheDir, presetName), nil
		}
		return "", fmt.Errorf("invalid input source type: %s", inputConfig.Type)
	default:
		return "", fmt.Errorf("unsupported input source type: %s", inputConfig.Type)
	}
}
