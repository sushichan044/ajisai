package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sushichan044/ajisai/utils"
)

type jsonLoader struct{}

func newJSONLoader() formatLoader[serializableConfig] {
	return &jsonLoader{}
}

func (l *jsonLoader) Load(configPath string) (*Config, error) {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	if _, statErr := os.Stat(resolvedPath); statErr != nil {
		return nil, fmt.Errorf("failed to get config file %s: %w", resolvedPath, statErr)
	}

	body, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", resolvedPath, err)
	}

	var cfgData serializableConfig
	if jsonErr := json.Unmarshal(body, &cfgData); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %s: %w", resolvedPath, jsonErr)
	}

	return l.fromFormat(cfgData)
}

func (l *jsonLoader) Save(configPath string, cfg *Config) error {
	resolvedPath, pathErr := utils.ResolveAbsPath(configPath)
	if pathErr != nil {
		return fmt.Errorf("failed to resolve config path: %w", pathErr)
	}
	cfgData, convErr := l.toFormat(cfg)
	if convErr != nil {
		return fmt.Errorf("failed to convert config to serializable format: %w", convErr)
	}

	jsonData, marshalErr := json.MarshalIndent(cfgData, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", marshalErr)
	}

	if err := utils.AtomicWriteFile(resolvedPath, bytes.NewReader(jsonData)); err != nil {
		return fmt.Errorf("failed to save config file atomically: %w", err)
	}

	return nil
}

//gocognit:ignore
func (l *jsonLoader) toFormat(cfg *Config) (serializableConfig, error) {
	var serializableCfg serializableConfig

	if cfg.Settings != nil {
		serializableCfg.Settings = &serializableSettings{
			CacheDir:     cfg.Settings.CacheDir,
			Experimental: cfg.Settings.Experimental,
			Namespace:    cfg.Settings.Namespace,
		}
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
			}
		}
		if cfg.Workspace.Integrations != nil {
			workspace.Integrations = &serializableAgentIntegration{
				Cursor: &serializableCursorIntegration{Enabled: cfg.Workspace.Integrations.Cursor.Enabled},
				GitHubCopilot: &serializableGitHubCopilotIntegration{
					Enabled: cfg.Workspace.Integrations.GitHubCopilot.Enabled,
				},
				Windsurf: &serializableWindsurfIntegration{Enabled: cfg.Workspace.Integrations.Windsurf.Enabled},
			}
		}
		serializableCfg.Workspace = &workspace
	}

	return serializableCfg, nil
}

func (l *jsonLoader) fromFormat(cfg serializableConfig) (*Config, error) {
	var settings Settings
	if cfg.Settings != nil {
		settings.CacheDir = cfg.Settings.CacheDir
		settings.Experimental = cfg.Settings.Experimental
		settings.Namespace = cfg.Settings.Namespace
	}

	var workspace Workspace
	if cfg.Workspace != nil {
		workspace.Imports = make(map[string]ImportedPackage, len(cfg.Workspace.Imports))

		for name, imp := range cfg.Workspace.Imports {
			switch ImportType(imp.Type) {
			case ImportTypeLocal:
				workspace.Imports[name] = ImportedPackage{
					Type: ImportTypeLocal,
					Details: LocalImportDetails{
						Path: imp.Path,
					},
					Include: imp.Include,
				}
				continue
			case ImportTypeGit:
				workspace.Imports[name] = ImportedPackage{
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
		if cfg.Workspace.Integrations != nil {
			workspace.Integrations = &AgentIntegrations{
				Cursor:        &CursorIntegration{Enabled: cfg.Workspace.Integrations.Cursor.Enabled},
				GitHubCopilot: &GitHubCopilotIntegration{Enabled: cfg.Workspace.Integrations.GitHubCopilot.Enabled},
				Windsurf:      &WindsurfIntegration{Enabled: cfg.Workspace.Integrations.Windsurf.Enabled},
			}
		}
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
