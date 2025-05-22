package config

import "fmt"

type (
	SerializableConfig struct {
		Settings  *serializableSettings  `json:"settings,omitempty"  yaml:"settings,omitempty"`
		Package   *serializablePackage   `json:"package,omitempty"   yaml:"package,omitempty"`
		Workspace *serializableWorkspace `json:"workspace,omitempty" yaml:"workspace,omitempty"`
	}

	serializableSettings struct {
		CacheDir     string `json:"cacheDir,omitempty"  yaml:"cacheDir,omitempty"`
		Experimental bool   `json:"experimental"        yaml:"experimental"`
		Namespace    string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
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
		Enabled bool `json:"enabled" yaml:"enabled"`
	}

	serializableGitHubCopilotIntegration struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	}

	serializableWindsurfIntegration struct {
		Enabled bool `json:"enabled" yaml:"enabled"`
	}
)

type configSerializerImpl struct{}

func NewSerializer() configSerializer {
	return &configSerializerImpl{}
}

func (s *configSerializerImpl) Serialize(cfg *Config) (SerializableConfig, error) {
	var serializableCfg SerializableConfig

	if cfg.Settings != nil {
		var settings serializableSettings
		settings = serializableSettings{
			CacheDir:     cfg.Settings.CacheDir,
			Experimental: cfg.Settings.Experimental,
			Namespace:    cfg.Settings.Namespace,
		}
		serializableCfg.Settings = &settings
	}

	if cfg.Package != nil {
		var pkg serializablePackage
		pkg.Name = cfg.Package.Name

		if cfg.Package.Exports != nil {
			pkg.Exports = make(map[string]serializableExportedPresetDefinition, len(cfg.Package.Exports))

			for name, export := range cfg.Package.Exports {
				pkg.Exports[name] = serializableExportedPresetDefinition(export)
			}
		}
		serializableCfg.Package = &pkg
	}

	if cfg.Workspace != nil {
		var workspace serializableWorkspace
		workspace.Imports = make(map[string]serializableImportedPackage, len(cfg.Workspace.Imports))

		for name, imp := range cfg.Workspace.Imports {
			switch imp.Type {
			case ImportTypeLocal:
				if details, ok := GetImportDetails[LocalImportDetails](imp); ok {
					workspace.Imports[name] = serializableImportedPackage{
						Type:    string(imp.Type),
						Path:    details.Path,
						Include: imp.Include,
					}
				}
			case ImportTypeGit:
				if details, ok := GetImportDetails[GitImportDetails](imp); ok {
					workspace.Imports[name] = serializableImportedPackage{
						Type:       string(imp.Type),
						Repository: details.Repository,
						Revision:   details.Revision,
						Include:    imp.Include,
					}
				}
			default:
				continue
				// TODO: Gracefully handling unsupported import type
				// return SerializableConfig{}, fmt.Errorf("unsupported import type: %s", imp.Type)
			}
		}

		if cfg.Workspace.Integrations != nil {
			var integrations serializableAgentIntegration

			if cfg.Workspace.Integrations.Cursor != nil {
				var cursor serializableCursorIntegration
				cursor = serializableCursorIntegration{
					Enabled: cfg.Workspace.Integrations.Cursor.Enabled,
				}
				integrations.Cursor = &cursor
			}

			if cfg.Workspace.Integrations.GitHubCopilot != nil {
				var githubCopilot serializableGitHubCopilotIntegration
				githubCopilot = serializableGitHubCopilotIntegration{
					Enabled: cfg.Workspace.Integrations.GitHubCopilot.Enabled,
				}
				integrations.GitHubCopilot = &githubCopilot
			}

			if cfg.Workspace.Integrations.Windsurf != nil {
				var windsurf serializableWindsurfIntegration
				windsurf = serializableWindsurfIntegration{
					Enabled: cfg.Workspace.Integrations.Windsurf.Enabled,
				}
				integrations.Windsurf = &windsurf
			}

			workspace.Integrations = &integrations
		}

		serializableCfg.Workspace = &workspace
	}

	return serializableCfg, nil
}

func (s *configSerializerImpl) Deserialize(cfg SerializableConfig) (*Config, error) {
	var settings Settings
	if cfg.Settings != nil {
		settings.CacheDir = cfg.Settings.CacheDir
		settings.Experimental = cfg.Settings.Experimental
		settings.Namespace = cfg.Settings.Namespace
	}

	var workspace Workspace
	if cfg.Workspace != nil {
		imports := make(map[string]ImportedPackage, len(cfg.Workspace.Imports))

		for name, imp := range cfg.Workspace.Imports {
			switch ImportType(imp.Type) {
			case ImportTypeLocal:
				imports[name] = ImportedPackage{
					Type: ImportTypeLocal,
					Details: LocalImportDetails{
						Path: imp.Path,
					},
					Include: imp.Include,
				}
				continue
			case ImportTypeGit:
				imports[name] = ImportedPackage{
					Type: ImportTypeGit,
					Details: GitImportDetails{
						Repository: imp.Repository,
						Revision:   imp.Revision,
					},
					Include: imp.Include,
				}
				continue
			}

			return nil, fmt.Errorf("unsupported import type: %s", imp.Type)
		}
		workspace.Imports = imports

		var integrations AgentIntegrations
		if cfg.Workspace.Integrations != nil {
			integrations = AgentIntegrations{}

			var cursor CursorIntegration
			if cfg.Workspace.Integrations.Cursor != nil {
				cursor = CursorIntegration{Enabled: cfg.Workspace.Integrations.Cursor.Enabled}
			}
			integrations.Cursor = &cursor

			var githubCopilot GitHubCopilotIntegration
			if cfg.Workspace.Integrations.GitHubCopilot != nil {
				githubCopilot = GitHubCopilotIntegration{
					Enabled: cfg.Workspace.Integrations.GitHubCopilot.Enabled,
				}
			}
			integrations.GitHubCopilot = &githubCopilot

			var windsurf WindsurfIntegration
			if cfg.Workspace.Integrations.Windsurf != nil {
				windsurf = WindsurfIntegration{
					Enabled: cfg.Workspace.Integrations.Windsurf.Enabled,
				}
			}
			integrations.Windsurf = &windsurf
		}
		workspace.Integrations = &integrations
	}

	var pkg Package
	if cfg.Package != nil {
		pkg.Name = cfg.Package.Name
		if cfg.Package.Exports != nil {
			pkg.Exports = make(map[string]ExportedPresetDefinition, len(cfg.Package.Exports))

			for name, export := range cfg.Package.Exports {
				pkg.Exports[name] = ExportedPresetDefinition(export)
			}
		}
	}

	return &Config{
		Settings:  &settings,
		Package:   &pkg,
		Workspace: &workspace,
	}, nil
}
