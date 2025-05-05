package domain

// PresetPackage represents the entire content fetched and parsed from a single input source.
type PresetPackage struct {
	InputKey string        `json:"inputKey"` // Key identifying the input source
	Items    []*PresetItem `json:"items"`    // Changed to slice of pointers
	// IgnorePatterns []string     // Glob patterns loaded from .aiignore at package root (optional)
	// Other potential package-level settings
}
