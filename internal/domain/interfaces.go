package domain

import "github.com/sushichan044/ajisai/internal/config"

type (
	// PackageFetcher retrieves packages from a source and stores them in the destination directory.
	PackageFetcher interface {
		// Fetch retrieves packages from the source and stores them in the destination directory.
		Fetch(source config.ImportedPackage, destinationDir string) error
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

	// PresetRepository is a repository for read / write Preset into a specific agent format.
	PresetRepository interface {
		// ReadPreset reads a preset from the given namespace.
		ReadPreset(namespace string) (AgentPreset, error)

		// WritePreset writes a preset to the given namespace.
		WritePreset(namespace string, preset AgentPreset) error

		// Clean removes all presets from the repository.
		Clean(namespace string) error
	}
)
