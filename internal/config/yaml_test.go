package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/config"
)

func TestYamlLoader_Load_FileNotFound(t *testing.T) {
	loader := config.NewYAMLLoader()
	_, err := loader.Load("non_existent_config.yaml")
	assert.Error(t, err)
}

func TestYamlLoader_Save_InvalidPath(t *testing.T) {
	loader := config.NewYAMLLoader()
	err := loader.Save("/invalid_path/should_not_exist/config.yaml", &config.Config{})
	assert.Error(t, err)
}

func TestYamlLoader_Load(t *testing.T) {
	tmp := t.TempDir()

	tcs := []struct {
		name     string
		yamlBody string
		expected *config.Config
	}{
		{
			name: "Full config",
			yamlBody: `
settings:
  cacheDir: /tmp/ajisai_cache
  experimental: true
  namespace: my_namespace
package:
  name: my_package
  exports:
    preset1:
      prompts:
        - prompts/prompt1.md
      rules:
        - rules/rule1.json
workspace:
  imports:
    import1:
      type: local
      path: path/to/import1
      include:
        - rule1
  integrations:
    cursor:
      enabled: true
    github-copilot:
      enabled: false
    windsurf:
      enabled: true
`,
			expected: &config.Config{
				Settings: &config.Settings{
					CacheDir:     "/tmp/ajisai_cache",
					Experimental: true,
					Namespace:    "my_namespace",
				},
				Package: &config.Package{
					Name: "my_package",
					Exports: map[string]config.ExportedPresetDefinition{
						"preset1": {Prompts: []string{"prompts/prompt1.md"}, Rules: []string{"rules/rule1.json"}},
					},
				},
				Workspace: &config.Workspace{
					Imports: map[string]config.ImportedPackage{
						"import1": {
							Type:    config.ImportTypeLocal,
							Details: config.LocalImportDetails{Path: "path/to/import1"},
							Include: []string{"rule1"},
						},
					},
					Integrations: &config.AgentIntegrations{
						Cursor:        &config.CursorIntegration{Enabled: true},
						GitHubCopilot: &config.GitHubCopilotIntegration{Enabled: false},
						Windsurf:      &config.WindsurfIntegration{Enabled: true},
					},
				},
			},
		},
		{
			name: "Empty config",
			yamlBody: `
`,
			expected: &config.Config{
				Settings:  &config.Settings{},
				Package:   &config.Package{},
				Workspace: &config.Workspace{},
			},
		},
		{
			name: "Partial Settings",
			yamlBody: `
settings:
  namespace: custom
  experimental: false
`,
			expected: &config.Config{
				Settings: &config.Settings{
					Namespace:    "custom",
					Experimental: false,
				},
				Package:   &config.Package{},
				Workspace: &config.Workspace{},
			},
		},
		{
			name: "Partial Package",
			yamlBody: `
package:
  name: my_package
`,
			expected: &config.Config{
				Settings:  &config.Settings{},
				Package:   &config.Package{Name: "my_package"},
				Workspace: &config.Workspace{},
			},
		},
		{
			name: "Partial Workspace: imports",
			yamlBody: `
workspace:
  imports:
    import1:
      type: local
      path: path/to/import1
`,
			expected: &config.Config{
				Settings: &config.Settings{},
				Package:  &config.Package{},
				Workspace: &config.Workspace{
					Imports: map[string]config.ImportedPackage{
						"import1": {
							Type:    config.ImportTypeLocal,
							Details: config.LocalImportDetails{Path: "path/to/import1"},
						},
					},
					Integrations: &config.AgentIntegrations{},
				},
			},
		},
		{
			name: "Partial Workspace: integrations",
			yamlBody: `
workspace:
  integrations:
    cursor:
      enabled: true
    github-copilot:
      enabled: false
`,
			expected: &config.Config{
				Settings: &config.Settings{},
				Package:  &config.Package{},
				Workspace: &config.Workspace{
					Imports: map[string]config.ImportedPackage{},
					Integrations: &config.AgentIntegrations{
						Cursor:        &config.CursorIntegration{Enabled: true},
						GitHubCopilot: &config.GitHubCopilotIntegration{Enabled: false},
						Windsurf:      &config.WindsurfIntegration{Enabled: false}, // zero value
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfgPath := filepath.Join(tmp, tc.name+".yaml")
			writeErr := os.WriteFile(cfgPath, []byte(tc.yamlBody), 0644)
			require.NoError(t, writeErr)

			loader := config.NewYAMLLoader()
			cfg, loadErr := loader.Load(cfgPath)
			require.NoError(t, loadErr)

			if res := cmp.Diff(tc.expected, cfg); res != "" {
				t.Errorf("mismatch (-want +got):\n%s", res)
			}
		})
	}
}

func TestYamlLoader_Save(t *testing.T) {
	tmp := t.TempDir()

	loader := config.NewYAMLLoader()

	tcs := []struct {
		name     string
		cfg      *config.Config
		expected string
	}{
		{
			name: "Full config",
			cfg: &config.Config{
				Settings: &config.Settings{
					CacheDir:     "/tmp/ajisai_cache",
					Experimental: true,
					Namespace:    "my_namespace",
				},
				Package: &config.Package{
					Name: "my_package",
					Exports: map[string]config.ExportedPresetDefinition{
						"preset1": {Prompts: []string{"prompts/prompt1.md"}, Rules: []string{"rules/rule1.json"}},
					},
				},
				Workspace: &config.Workspace{
					Imports: map[string]config.ImportedPackage{
						"import1": {
							Type:    config.ImportTypeLocal,
							Details: config.LocalImportDetails{Path: "path/to/import1"},
						},
					},
					Integrations: &config.AgentIntegrations{
						Cursor:        &config.CursorIntegration{Enabled: true},
						GitHubCopilot: &config.GitHubCopilotIntegration{Enabled: false},
						Windsurf:      &config.WindsurfIntegration{Enabled: true},
					},
				},
			},
			expected: `settings:
  cacheDir: /tmp/ajisai_cache
  experimental: true
  namespace: my_namespace
package:
  exports:
    preset1:
      prompts:
      - prompts/prompt1.md
      rules:
      - rules/rule1.json
  name: my_package
workspace:
  imports:
    import1:
      type: local
      path: path/to/import1
  integrations:
    cursor:
      enabled: true
    github-copilot:
      enabled: false
    windsurf:
      enabled: true
`,
		},
		{
			name: "Partial Settings",
			cfg: &config.Config{
				Settings: &config.Settings{
					Namespace:    "custom",
					Experimental: false,
				},
			},
			expected: `settings:
  experimental: false
  namespace: custom
`,
		},
		{
			name: "Partial Package",
			cfg: &config.Config{
				Package: &config.Package{Name: "my_package"},
			},
			expected: `package:
  name: my_package
`,
		},
		{
			name: "Partial Workspace: imports",
			cfg: &config.Config{
				Workspace: &config.Workspace{
					Imports: map[string]config.ImportedPackage{
						"import1": {
							Type:    config.ImportTypeLocal,
							Details: config.LocalImportDetails{Path: "path/to/import1"},
						},
					},
				},
			},
			expected: `workspace:
  imports:
    import1:
      type: local
      path: path/to/import1
`,
		},
		{
			name: "Partial Workspace: integrations",
			cfg: &config.Config{
				Workspace: &config.Workspace{
					Integrations: &config.AgentIntegrations{
						Cursor:        &config.CursorIntegration{Enabled: true},
						GitHubCopilot: &config.GitHubCopilotIntegration{Enabled: false},
					},
				},
			},
			expected: `workspace:
  integrations:
    cursor:
      enabled: true
    github-copilot:
      enabled: false
`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfgPath := filepath.Join(tmp, tc.name+".yaml")
			err := loader.Save(cfgPath, tc.cfg)
			require.NoError(t, err)

			actual, err := os.ReadFile(cfgPath)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(actual))
		})
	}
}
