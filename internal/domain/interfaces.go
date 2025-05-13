package domain

type (
	// ContentFetcher retrieves content from a source defined by InputSource.
	ContentFetcher interface {
		// Fetch retrieves content from the source and stores it in the destinationDir.
		Fetch(source InputSource, destinationDir string) error
	}

	// AgentBridge is a bridge between the domain and the agent.
	// It converts between the domain and the agent's format.
	AgentBridge[TRule any, TPrompt any] interface {
		ToAgentRule(rule RuleItem) (TRule, error)
		FromAgentRule(rule TRule) (RuleItem, error)

		SerializeAgentRule(rule TRule) (string, error)
		DeserializeAgentRule(slug string, ruleBody string) (TRule, error)

		FromAgentPrompt(prompt TPrompt) (PromptItem, error)
		ToAgentPrompt(prompt PromptItem) (TPrompt, error)

		SerializeAgentPrompt(prompt TPrompt) (string, error)
		DeserializeAgentPrompt(slug string, promptBody string) (TPrompt, error)
	}

	// PresetRepository is a repository for read / write PresetPackage into a specific agent format.
	PresetRepository interface {
		// ReadPackage reads a preset package from the given namespace.
		ReadPackage(namespace string) (PresetPackage, error)

		// WritePackage writes a preset package to the given namespace.
		WritePackage(namespace string, pkg PresetPackage) error

		// Clean removes all presets from the repository.
		Clean(namespace string) error
	}
)
