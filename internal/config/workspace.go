package config

const (
	ImportTypeLocal ImportType = "local" // Local file system import
	ImportTypeGit   ImportType = "git"   // Git repository import

	AgentIntegrationTypeCursor        AgentIntegrationType = "cursor"         // Cursor output target
	AgentIntegrationTypeGitHubCopilot AgentIntegrationType = "github-copilot" // GitHub Copilot output target
	AgentIntegrationTypeWindsurf      AgentIntegrationType = "windsurf"       // WindSurf output target
)

type (
	ImportType           string
	AgentIntegrationType string

	/*
		Workspace defines a workspace definition.
	*/
	Workspace struct {
		/*
			Imported preset packages to use in this workspace.

			ajisai will fetch the presets from the source and store them in the cache directory.

			Key is used to identify the preset package.
		*/
		Imports map[string]ImportedPackage `json:"imports,omitempty"`

		/*
			Agents to integrate imported presets into.
		*/
		Integrations *AgentIntegrations `json:"integrations,omitempty"`
	}

	// ImportedPackage defines a package that will be imported into the workspace.
	ImportedPackage struct {
		/*
			Type identifier (e.g., "local", "git").
		*/
		Type ImportType `json:"type"`

		/*
			List of exported presets to include in the workspace.
		*/
		Include []string `json:"include,omitempty"`

		/*
			Type-specific configuration details.
		*/
		Details ImportDetails `json:"details"`
	}

	// ImportDetails is an interface for type-specific input source configurations.
	ImportDetails interface {
		isImportDetails()
	}

	LocalImportDetails struct {
		Path string // Path to the local directory
	}

	// GitImportDetails holds configuration specific to Git repository inputs.
	GitImportDetails struct {
		Repository string // URL of the Git repository
		Revision   string // Optional branch, tag, or commit SHA (defaults to latest)
	}

	// AgentIntegrations defines specific integrations for each agent.
	AgentIntegrations struct {
		Cursor        *CursorIntegration        `json:"cursor,omitempty"`
		GitHubCopilot *GitHubCopilotIntegration `json:"github-copilot,omitempty"`
		Windsurf      *WindsurfIntegration      `json:"windsurf,omitempty"`
	}

	CursorIntegration struct {
		Enabled bool `json:"enabled,omitempty"`
	}

	GitHubCopilotIntegration struct {
		Enabled bool `json:"enabled,omitempty"`
	}

	WindsurfIntegration struct {
		Enabled bool `json:"enabled,omitempty"`
	}
)

// GetImportDetails safely performs a type assertion on UsingPresetPackageSource.Details.
func GetImportDetails[T ImportDetails](is ImportedPackage) (T, bool) {
	details, ok := is.Details.(T)
	return details, ok
}

func (d LocalImportDetails) isImportDetails() {}

func (d GitImportDetails) isImportDetails() {}

func applyDefaultsToWorkspace(workspace *Workspace) *Workspace {
	if workspace == nil {
		workspace = &Workspace{}
	}

	if workspace.Imports == nil {
		workspace.Imports = map[string]ImportedPackage{}
	}

	workspace.Integrations = applyDefaultsToAgentIntegrations(workspace.Integrations)

	return workspace
}

func applyDefaultsToAgentIntegrations(integrations *AgentIntegrations) *AgentIntegrations {
	if integrations == nil {
		integrations = &AgentIntegrations{}
	}

	if integrations.Cursor == nil {
		integrations.Cursor = &CursorIntegration{}
	}

	if integrations.GitHubCopilot == nil {
		integrations.GitHubCopilot = &GitHubCopilotIntegration{}
	}

	if integrations.Windsurf == nil {
		integrations.Windsurf = &WindsurfIntegration{}
	}

	return integrations
}
