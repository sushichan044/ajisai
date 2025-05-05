package domain

// PresetPackage represents the entire content fetched and parsed from a single input source.
type PresetPackage struct {
	InputKey string       // Identifier from [inputs] section (e.g., "common-prompts")
	Items    []PresetItem // List of rules and prompts in this package
	// IgnorePatterns []string     // Glob patterns loaded from .aiignore at package root (optional)
	// Other potential package-level settings
}
