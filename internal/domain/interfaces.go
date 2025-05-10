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

// AgentBridge is a bridge between the domain and the agent.
// It converts between the domain and the agent's format.
type AgentBridge[TRule any, TPrompt any] interface {
	ToAgentRule(rule RuleItem) (TRule, error)
	FromAgentRule(rule TRule) (RuleItem, error)

	FromAgentPrompt(prompt TPrompt) (PromptItem, error)
	ToAgentPrompt(prompt PromptItem) (TPrompt, error)
}

// PresetRepository is a repository for read / write PresetPackage into a specific agent format.
type PresetRepository interface {
	// ReadPackage reads a preset package from the given namespace.
	ReadPackage(namespace string) (PresetPackage, error)

	// WritePackage writes a preset package to the given namespace.
	WritePackage(namespace string, pkg PresetPackage) error
}
