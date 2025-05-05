package domain

// PresetItem represents a single rule or prompt file within a PresetPackage.
type PresetItem struct {
	Name         string      `json:"name"`         // Unique name within package (e.g., "my-rule")
	Description  string      `json:"description"`  // Content (Markdown), excluding front matter
	Type         string      `json:"type"`         // "rule" or "prompt"
	RelativePath string      `json:"relativePath"` // Path relative to package root (e.g., "rules/my-rule.md")
	Metadata     interface{} `json:"metadata"`     // Changed to interface{}
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
