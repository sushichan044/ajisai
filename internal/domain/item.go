package domain

// PresetType defines the kind of preset (e.g., rule, prompt).
type PresetType string

const (
	RulePresetType   PresetType = "rule"
	PromptPresetType PresetType = "prompt"
)

// PresetItem represents a single rule or prompt file within a PresetPackage.
type PresetItem struct {
	Name         string     // Unique name within package (e.g., "my-rule")
	Content      string     // Content (e.g., Markdown), excluding front matter
	Type         PresetType // "rule" or "prompt"
	RelativePath string     // Path relative to package root (e.g., "rules/my-rule.md")
	Metadata     any        // Decoded front matter (e.g., RuleMetadata, PromptMetadata)
}

// RuleMetadata defines the structure for metadata specific to rules.
type RuleMetadata struct {
	Title       string   `mapstructure:"title,omitempty"`       // Optional: User-facing title from front matter.
	Description string   `mapstructure:"description,omitempty"` // Optional: Detailed description from front matter.
	Attach      string   `mapstructure:"attach"`                // Required: How the rule is attached ("always", "glob", "manual", etc.). No default value.
	Glob        []string `mapstructure:"glob,omitempty"`        // Optional: Glob patterns, used when Attach is "glob".
}

// PromptMetadata defines the structure for metadata specific to prompts.
type PromptMetadata struct {
	Description string `mapstructure:"description,omitempty"` // Optional: Detailed description from front matter.
}
