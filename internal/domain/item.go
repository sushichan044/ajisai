package domain

// PresetItem represents a single rule or prompt file within a PresetPackage.
type PresetItem struct {
	Name         string                 `json:"name"`               // Unique name within package (e.g., "my-rule")
	Description  string                 `json:"description"`        // Content (Markdown), excluding front matter
	Type         string                 `json:"type"`               // "rule" or "prompt"
	RelativePath string                 `json:"relativePath"`       // Path relative to package root (e.g., "rules/my-rule.md")
	Metadata     map[string]interface{} `json:"metadata,omitempty"` // Data parsed from YAML front matter (primarily for rules)
}
