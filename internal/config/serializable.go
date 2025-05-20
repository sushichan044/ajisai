package config

type (
	serializableConfig struct {
		Settings  *serializableSettings  `json:"settings,omitempty"  yaml:"settings,omitempty"`
		Package   *serializablePackage   `json:"package,omitempty"   yaml:"package,omitempty"`
		Workspace *serializableWorkspace `json:"workspace,omitempty" yaml:"workspace,omitempty"`
	}

	serializableSettings struct {
		CacheDir     string `json:"cacheDir,omitempty"     yaml:"cacheDir,omitempty"`
		Experimental bool   `json:"experimental,omitempty" yaml:"experimental,omitempty"`
		Namespace    string `json:"namespace,omitempty"    yaml:"namespace,omitempty"`
	}

	serializablePackage struct {
		Exports map[string]serializableExportedPresetDefinition `json:"exports,omitempty" yaml:"exports,omitempty"`
		Name    string                                          `json:"name"              yaml:"name"`
	}

	serializableExportedPresetDefinition struct {
		Prompts []string `json:"prompts,omitempty" yaml:"prompts,omitempty"`
		Rules   []string `json:"rules,omitempty"   yaml:"rules,omitempty"`
	}

	serializableWorkspace struct {
		Imports      map[string]serializableImportedPackage `json:"imports,omitempty"      yaml:"imports,omitempty"`
		Integrations *serializableAgentIntegration          `json:"integrations,omitempty" yaml:"integrations,omitempty"`
	}

	serializableImportedPackage struct {
		Type       string   `json:"type"                 yaml:"type"`
		Include    []string `json:"include,omitempty"    yaml:"include,omitempty"`
		Path       string   `json:"path,omitempty"       yaml:"path,omitempty"`       // only for type: local
		Repository string   `json:"repository,omitempty" yaml:"repository,omitempty"` // only for type: git
		Revision   string   `json:"revision,omitempty"   yaml:"revision,omitempty"`   // only for type: git
	}

	serializableAgentIntegration struct {
		Cursor        *serializableCursorIntegration        `json:"cursor,omitempty"         yaml:"cursor,omitempty"`
		GitHubCopilot *serializableGitHubCopilotIntegration `json:"github-copilot,omitempty" yaml:"github-copilot,omitempty"`
		Windsurf      *serializableWindsurfIntegration      `json:"windsurf,omitempty"       yaml:"windsurf,omitempty"`
	}

	serializableCursorIntegration struct {
		Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	}

	serializableGitHubCopilotIntegration struct {
		Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	}

	serializableWindsurfIntegration struct {
		Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	}
)
