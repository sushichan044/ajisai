package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sushichan044/ajisai/utils"
)

type jsonLoader struct{}

func NewJSONLoader() formatLoader[jsonConfig] {
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

	var jsonCfg jsonConfig
	if jsonErr := json.Unmarshal(body, &jsonCfg); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %s: %w", resolvedPath, jsonErr)
	}

	return l.fromFormat(jsonCfg), nil
}

func (l *jsonLoader) Save(configPath string, cfg *Config) error {
	resolvedPath, err := utils.ResolveAbsPath(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}
	jsonCfg := l.toFormat(cfg)

	tempFile, tempErr := os.CreateTemp("", "ajisai-*.json.tmp")
	if tempErr != nil {
		return fmt.Errorf("failed to create temporary file: %w", tempErr)
	}

	tmpFileName := tempFile.Name()
	defer os.Remove(tmpFileName)
	defer tempFile.Close()

	encoder := json.NewEncoder(tempFile)
	if encodeErr := encoder.Encode(jsonCfg); encodeErr != nil {
		return fmt.Errorf("failed to encode config: %w", encodeErr)
	}

	if renameErr := os.Rename(tmpFileName, resolvedPath); renameErr != nil {
		return fmt.Errorf("failed to rename temporary file to target file %s: %w", resolvedPath, renameErr)
	}

	return nil
}

//gocognit:ignore
func (l *jsonLoader) toFormat(cfg *Config) jsonConfig {
	var jsonCfg jsonConfig

	if cfg.Settings != nil {
		jsonCfg.Settings = &jsonSettings{
			CacheDir:     cfg.Settings.CacheDir,
			Experimental: cfg.Settings.Experimental,
			Namespace:    cfg.Settings.Namespace,
		}
	}

	if cfg.Package != nil {
		var pkg jsonPackage
		pkg.Name = cfg.Package.Name
		if cfg.Package.Exports != nil {
			pkg.Exports = make(map[string]jsonExportedPresetDefinition)
			for name, export := range cfg.Package.Exports {
				pkg.Exports[name] = jsonExportedPresetDefinition(export)
			}
		}
		jsonCfg.Package = &pkg
	}

	if cfg.Workspace != nil {
		var workspace jsonWorkspace
		workspace.Imports = make(map[string]jsonImportedPackage)
		for name, imp := range cfg.Workspace.Imports {
			switch imp.Type {
			case ImportTypeLocal:
				if details, ok := GetImportDetails[LocalImportDetails](imp); ok {
					workspace.Imports[name] = jsonImportedPackage{
						Type:    string(imp.Type),
						Path:    details.Path,
						Include: imp.Include,
					}
				}
			case ImportTypeGit:
				if details, ok := GetImportDetails[GitImportDetails](imp); ok {
					workspace.Imports[name] = jsonImportedPackage{
						Type:       string(imp.Type),
						Repository: details.Repository,
						Revision:   details.Revision,
						Include:    imp.Include,
					}
				}
			}
		}
		if cfg.Workspace.Integrations != nil {
			workspace.Integrations = &jsonAgentIntegration{
				Cursor:        &jsonCursorIntegration{Enabled: cfg.Workspace.Integrations.Cursor.Enabled},
				GitHubCopilot: &jsonGitHubCopilotIntegration{Enabled: cfg.Workspace.Integrations.GitHubCopilot.Enabled},
				Windsurf:      &jsonWindsurfIntegration{Enabled: cfg.Workspace.Integrations.Windsurf.Enabled},
			}
		}
		jsonCfg.Workspace = &workspace
	}

	return jsonCfg
}

func (l *jsonLoader) fromFormat(cfg jsonConfig) *Config {
	var settings Settings
	if cfg.Settings != nil {
		settings.CacheDir = cfg.Settings.CacheDir
		settings.Experimental = cfg.Settings.Experimental
		settings.Namespace = cfg.Settings.Namespace
	}

	var workspace Workspace
	workspace.Imports = make(map[string]ImportedPackage)
	if cfg.Workspace != nil {
		for name, imp := range cfg.Workspace.Imports {
			switch imp.Type {
			case string(ImportTypeLocal):
				workspace.Imports[name] = ImportedPackage{
					Type: ImportTypeLocal,
					Details: LocalImportDetails{
						Path: imp.Path,
					},
					Include: imp.Include,
				}
			case string(ImportTypeGit):
				workspace.Imports[name] = ImportedPackage{
					Type: ImportTypeGit,
					Details: GitImportDetails{
						Repository: imp.Repository,
						Revision:   imp.Revision,
					},
					Include: imp.Include,
				}
			}
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
			pkg.Exports = make(map[string]ExportedPresetDefinition)
			for name, export := range cfg.Package.Exports {
				pkg.Exports[name] = ExportedPresetDefinition(export)
			}
		}
	}

	return &Config{
		Settings:  &settings,
		Package:   &pkg,
		Workspace: &workspace,
	}
}

type (
	jsonConfig struct {
		Settings  *jsonSettings  `json:"settings,omitempty"`
		Package   *jsonPackage   `json:"package,omitempty"`
		Workspace *jsonWorkspace `json:"workspace,omitempty"`
	}

	jsonSettings struct {
		CacheDir     string `json:"cacheDir,omitempty"`
		Experimental bool   `json:"experimental,omitempty"`
		Namespace    string `json:"namespace,omitempty"`
	}

	jsonPackage struct {
		Exports map[string]jsonExportedPresetDefinition `json:"exports,omitempty"`
		Name    string                                  `json:"name"`
	}

	jsonExportedPresetDefinition struct {
		Prompts []string `json:"prompts,omitempty"`
		Rules   []string `json:"rules,omitempty"`
	}

	jsonWorkspace struct {
		Imports      map[string]jsonImportedPackage `json:"imports,omitempty"`
		Integrations *jsonAgentIntegration          `json:"integrations,omitempty"`
	}

	jsonImportedPackage struct {
		Type       string   `json:"type"`
		Include    []string `json:"include,omitempty"`
		Path       string   `json:"path,omitempty"`       // only for type: local
		Repository string   `json:"repository,omitempty"` // only for type: git
		Revision   string   `json:"revision,omitempty"`   // only for type: git
	}

	jsonAgentIntegration struct {
		Cursor        *jsonCursorIntegration        `json:"cursor,omitempty"`
		GitHubCopilot *jsonGitHubCopilotIntegration `json:"github-copilot,omitempty"`
		Windsurf      *jsonWindsurfIntegration      `json:"windsurf,omitempty"`
	}

	jsonCursorIntegration struct {
		Enabled bool `json:"enabled,omitempty"`
	}

	jsonGitHubCopilotIntegration struct {
		Enabled bool `json:"enabled,omitempty"`
	}

	jsonWindsurfIntegration struct {
		Enabled bool `json:"enabled,omitempty"`
	}
)
