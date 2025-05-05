package domain

import (
	"context"
)

// ConfigManager handles loading and saving of application configuration.
type ConfigManager interface {
	// Load reads the configuration file from the given path,
	// unmarshals it into the internal Config struct (handling InputSource types),
	// validates it, and applies defaults.
	Load(configPath string) (*Config, error)

	// Save writes the given internal configuration representation
	// back to the specified file path.
	// Note: Saving might lose comments/formatting from the original TOML.
	Save(configPath string, cfg *Config) error
}

// ContentFetcher retrieves content from a source defined by InputSource.
type ContentFetcher interface {
	// Fetch retrieves content from the source and stores it in the destinationDir.
	Fetch(ctx context.Context, source InputSource, destinationDir string) error
}

// ConfigPresetParser parses content from a source directory into a PresetPackage.
type ConfigPresetParser interface {
	// Parse reads the content of sourceDir, interprets it based on a specific format,
	// and returns a structured PresetPackage.
	Parse(ctx context.Context, inputKey string, sourceDir string) (*PresetPackage, error)
}

// AIAgentConfigurationAdapter writes the collected presets to a target AI agent's format.
type AIAgentConfigurationAdapter interface {
	// Write takes the collected packages (map key is the input source key)
	// and writes them out in the format required by the specific AI agent,
	// potentially organizing them under the given namespace.
	Write(ctx context.Context, namespace string, packages map[string]*PresetPackage) error
}

// --- Potentially add interfaces for 'import' and 'doctor' command helpers later ---

// DefaultFormatWriter (Conceptual for 'import').
type DefaultFormatWriter interface {
	Write(ctx context.Context, items []PresetItem, outputDir string) error
}

// DefaultFormatValidator (Conceptual for 'doctor').
type ValidationIssue struct {
	Path     string
	Severity string // e.g., "error", "warning"
	Message  string
}

type DefaultFormatValidator interface {
	Validate(ctx context.Context, targetDir string) ([]ValidationIssue, error)
}
