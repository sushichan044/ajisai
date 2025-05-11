package domain

const (
	InputSourceTypeLocal InputSourceType = "local" // Local file system input
	InputSourceTypeGit   InputSourceType = "git"   // Git repository input

	OutputTargetTypeCursor        OutputTargetType = "cursor"         // Cursor output target
	OutputTargetTypeGitHubCopilot OutputTargetType = "github-copilot" // GitHub Copilot output target
)

type (
	InputSourceType  string
	OutputTargetType string

	// Config represents the fully resolved and validated application configuration.
	Config struct {
		Settings                         // Resolved top-level settings
		Inputs   map[string]InputSource  // Key is the input source identifier
		Outputs  map[string]OutputTarget // Key is the output target identifier
	}

	// Settings holds application-wide settings with defaults applied.
	Settings struct {
		CacheDir  string // Resolved cache directory path
		Namespace string // Resolved namespace
	}

	// InputSource defines a configured source for presets.
	InputSource struct {
		Type    InputSourceType    // Type identifier (e.g., "local", "git")
		Details InputSourceDetails // Type-specific configuration details
	}

	// InputSourceDetails is an interface for type-specific input source configurations.
	InputSourceDetails interface {
		isInputSourceDetails()
	}

	LocalInputSourceDetails struct {
		Path string // Path to the local directory
	}

	// GitInputSourceDetails holds configuration specific to Git repository inputs.
	GitInputSourceDetails struct {
		Repository string // URL of the Git repository
		Revision   string // Optional branch, tag, or commit SHA (defaults resolved by Fetcher)
		Directory  string // Optional subdirectory within the repo
	}

	// OutputTarget defines a configured destination for the processed presets.
	OutputTarget struct {
		Target  OutputTargetType // Type of output target (e.g., "cursor", "github-copilot")
		Enabled bool
	}
)

func (d LocalInputSourceDetails) isInputSourceDetails() {}

func (d GitInputSourceDetails) isInputSourceDetails() {}

// GetInputSourceDetails safely performs a type assertion on InputSource.Details.
func GetInputSourceDetails[T InputSourceDetails](is InputSource) (T, bool) {
	details, ok := is.Details.(T)
	return details, ok
}
