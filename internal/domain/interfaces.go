package domain

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
	Fetch(source InputSource, destinationDir string) error
}

type AgentAdapter[TRule any, TPrompt any] interface {
	ToAgentRule(rule RuleItem) (TRule, error)
	FromAgentRule(rule TRule) (RuleItem, error)

	FromAgentPrompt(prompt TPrompt) (PromptItem, error)
	ToAgentPrompt(prompt PromptItem) (TPrompt, error)

	WritePackage(namespace string, pkg PresetPackage) error
	ReadPackage(namespace string, pkg PresetPackage) error
}
