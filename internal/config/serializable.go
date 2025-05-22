package config

import (
	"errors"
	"fmt"
)

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

type recoverableNilInputError struct {
}

func (e *recoverableNilInputError) Error() string {
	return "sentinel: this is valid early return"
}

func (e *recoverableNilInputError) Unwrap() error {
	return nil
}

func (s *configSerializerImpl) Serialize(cfg *Config) (SerializableConfig, error) {
	serializableSettings := serializeSettings(cfg.Settings)
	serializablePackage := serializePackage(cfg.Package)
	serializableWorkspace, err := serializeWorkspace(cfg.Workspace)

	if err != nil && !errors.Is(err, &recoverableNilInputError{}) {
		return SerializableConfig{}, err
	}

	return SerializableConfig{
		Settings:  serializableSettings,
		Package:   serializablePackage,
		Workspace: serializableWorkspace,
	}, nil
}

func (s *configSerializerImpl) Deserialize(cfg SerializableConfig) (*Config, error) {
	settings := deserializeSettings(cfg.Settings)
	pkg := deserializePackage(cfg.Package)
	workspace, err := deserializeWorkspace(cfg.Workspace)
	if err != nil {
		return nil, err
	}

	return &Config{
		Settings:  settings,
		Package:   pkg,
		Workspace: workspace,
	}, nil
}

func serializeSettings(settings *Settings) *serializableSettings {
	if settings == nil {
		return nil
	}

	return &serializableSettings{
		CacheDir:     settings.CacheDir,
		Experimental: settings.Experimental,
		Namespace:    settings.Namespace,
	}
}

func deserializeSettings(serializableSettings *serializableSettings) *Settings {
	var settings Settings

	if serializableSettings == nil {
		return &settings
	}

	settings.CacheDir = serializableSettings.CacheDir
	settings.Experimental = serializableSettings.Experimental
	settings.Namespace = serializableSettings.Namespace

	return &settings
}

func serializePackage(pkg *Package) *serializablePackage {
	if pkg == nil {
		return nil
	}

	var s serializablePackage
	s.Name = pkg.Name

	if pkg.Exports != nil {
		s.Exports = make(map[string]serializableExportedPresetDefinition, len(pkg.Exports))

		for name, export := range pkg.Exports {
			s.Exports[name] = serializableExportedPresetDefinition(export)
		}
	}

	return &s
}

func deserializePackage(sPkg *serializablePackage) *Package {
	var pkg Package
	if sPkg != nil {
		pkg.Name = sPkg.Name
		if sPkg.Exports != nil {
			pkg.Exports = make(map[string]ExportedPresetDefinition, len(sPkg.Exports))

			for name, export := range sPkg.Exports {
				pkg.Exports[name] = ExportedPresetDefinition(export)
			}
		}
	}

	return &pkg
}

func serializeWorkspace(workspace *Workspace) (*serializableWorkspace, error) {
	if workspace == nil {
		// sentinel: this is valid early return
		return nil, &recoverableNilInputError{}
	}

	var s serializableWorkspace
	s.Imports = make(map[string]serializableImportedPackage, len(workspace.Imports))

	for name, imp := range workspace.Imports {
		switch imp.Type {
		case ImportTypeLocal:
			if details, ok := GetImportDetails[LocalImportDetails](imp); ok {
				s.Imports[name] = serializableImportedPackage{
					Type:    string(imp.Type),
					Path:    details.Path,
					Include: imp.Include,
				}
			}
		case ImportTypeGit:
			if details, ok := GetImportDetails[GitImportDetails](imp); ok {
				s.Imports[name] = serializableImportedPackage{
					Type:       string(imp.Type),
					Repository: details.Repository,
					Revision:   details.Revision,
					Include:    imp.Include,
				}
			}
		default:
			return nil, fmt.Errorf("unsupported import type: %s", imp.Type)
		}
	}

	if workspace.Integrations != nil {
		var integrations serializableAgentIntegration

		if workspace.Integrations.Cursor != nil {
			var cursor = serializableCursorIntegration{
				Enabled: workspace.Integrations.Cursor.Enabled,
			}
			integrations.Cursor = &cursor
		}

		if workspace.Integrations.GitHubCopilot != nil {
			var githubCopilot = serializableGitHubCopilotIntegration{
				Enabled: workspace.Integrations.GitHubCopilot.Enabled,
			}
			integrations.GitHubCopilot = &githubCopilot
		}

		if workspace.Integrations.Windsurf != nil {
			var windsurf = serializableWindsurfIntegration{
				Enabled: workspace.Integrations.Windsurf.Enabled,
			}
			integrations.Windsurf = &windsurf
		}

		s.Integrations = &integrations
	}

	return &s, nil
}

func deserializeWorkspace(sWorkspace *serializableWorkspace) (*Workspace, error) {
	var workspace Workspace
	if sWorkspace == nil {
		return &workspace, nil
	}

	imports := make(map[string]ImportedPackage, len(sWorkspace.Imports))
	for name, imp := range sWorkspace.Imports {
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
	if sWorkspace.Integrations != nil {
		integrations = AgentIntegrations{}

		var cursor CursorIntegration
		if sWorkspace.Integrations.Cursor != nil {
			cursor = CursorIntegration{Enabled: sWorkspace.Integrations.Cursor.Enabled}
		}
		integrations.Cursor = &cursor

		var githubCopilot GitHubCopilotIntegration
		if sWorkspace.Integrations.GitHubCopilot != nil {
			githubCopilot = GitHubCopilotIntegration{
				Enabled: sWorkspace.Integrations.GitHubCopilot.Enabled,
			}
		}
		integrations.GitHubCopilot = &githubCopilot

		var windsurf WindsurfIntegration
		if sWorkspace.Integrations.Windsurf != nil {
			windsurf = WindsurfIntegration{
				Enabled: sWorkspace.Integrations.Windsurf.Enabled,
			}
		}
		integrations.Windsurf = &windsurf
	}
	workspace.Integrations = &integrations

	return &workspace, nil
}
