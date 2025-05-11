package domain

// Config represents the fully resolved and validated application configuration.
type Config struct {
	Global  GlobalConfig            // Resolved global settings
	Inputs  map[string]InputSource  // Key is the input source identifier
	Outputs map[string]OutputTarget // Key is the output target identifier
}

// GlobalConfig holds application-wide settings with defaults applied.
type GlobalConfig struct {
	CacheDir  string // Resolved cache directory path
	Namespace string // Resolved namespace
}

// --- InputSource Definitions ---

// InputSource defines a configured source for presets.
type InputSource struct {
	Type    string             // Type identifier (e.g., "local", "git")
	Details InputSourceDetails // Type-specific configuration details
}

// InputSourceDetails is an interface for type-specific input source configurations.
type InputSourceDetails interface {
	isInputSourceDetails()
}

// LocalInputSourceDetails holds configuration specific to local directory inputs.
type LocalInputSourceDetails struct {
	Path string // Path to the local directory
}

func (d LocalInputSourceDetails) isInputSourceDetails() {}

// GitInputSourceDetails holds configuration specific to Git repository inputs.
type GitInputSourceDetails struct {
	Repository string // URL of the Git repository
	Revision   string // Optional branch, tag, or commit SHA (defaults resolved by Fetcher)
	Directory  string // Optional subdirectory within the repo
}

func (d GitInputSourceDetails) isInputSourceDetails() {}

// OutputTarget defines a configured destination for the processed presets.
type OutputTarget struct {
	Target  string // Type identifier (e.g., "cursor", "vscode-copilot")
	Enabled bool   // Whether this output target is active (default: true)
}

// GetInputSourceDetails safely performs a type assertion on InputSource.Details.
func GetInputSourceDetails[T InputSourceDetails](is InputSource) (T, bool) {
	details, ok := is.Details.(T)
	return details, ok
}
