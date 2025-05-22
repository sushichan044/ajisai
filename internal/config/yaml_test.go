package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYamlLoader_Load_FileNotFound(t *testing.T) {
	loader := newYAMLLoader()
	_, err := loader.Load("non_existent_config.yaml")
	assert.Error(t, err)
}

func TestYamlLoader_Save_InvalidPath(t *testing.T) {
	loader := newYAMLLoader()
	err := loader.Save("/invalid_path/should_not_exist/config.yaml", &Config{})
	assert.Error(t, err)
}

func TestYamlLoader_Load(t *testing.T) {
	tmp := t.TempDir()

	tcs := []struct {
		name     string
		yamlBody string
		expected *Config
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
			expected: &Config{
				Settings: &Settings{
					CacheDir:     "/tmp/ajisai_cache",
					Experimental: true,
					Namespace:    "my_namespace",
				},
				Package: &Package{
					Name: "my_package",
					Exports: map[string]ExportedPresetDefinition{
						"preset1": {Prompts: []string{"prompts/prompt1.md"}, Rules: []string{"rules/rule1.json"}},
					},
				},
				Workspace: &Workspace{
					Imports: map[string]ImportedPackage{
						"import1": {
							Type:    ImportTypeLocal,
							Details: LocalImportDetails{Path: "path/to/import1"},
							Include: []string{"rule1"},
						},
					},
					Integrations: &AgentIntegrations{
						Cursor:        &CursorIntegration{Enabled: true},
						GitHubCopilot: &GitHubCopilotIntegration{Enabled: false},
						Windsurf:      &WindsurfIntegration{Enabled: true},
					},
				},
			},
		},
		{
			name: "Empty config",
			yamlBody: `
`,
			expected: &Config{
				Settings:  &Settings{},
				Package:   &Package{},
				Workspace: &Workspace{},
			},
		},
		{
			name: "Partial Settings",
			yamlBody: `
settings:
  namespace: custom
  experimental: false
`,
			expected: &Config{
				Settings: &Settings{
					Namespace:    "custom",
					Experimental: false,
				},
				Package:   &Package{},
				Workspace: &Workspace{},
			},
		},
		{
			name: "Partial Package",
			yamlBody: `
package:
  name: my_package
`,
			expected: &Config{
				Settings:  &Settings{},
				Package:   &Package{Name: "my_package"},
				Workspace: &Workspace{},
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
			expected: &Config{
				Settings: &Settings{},
				Package:  &Package{},
				Workspace: &Workspace{
					Imports: map[string]ImportedPackage{
						"import1": {Type: ImportTypeLocal, Details: LocalImportDetails{Path: "path/to/import1"}},
					},
					Integrations: &AgentIntegrations{},
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
			expected: &Config{
				Settings: &Settings{},
				Package:  &Package{},
				Workspace: &Workspace{
					Imports: map[string]ImportedPackage{},
					Integrations: &AgentIntegrations{
						Cursor:        &CursorIntegration{Enabled: true},
						GitHubCopilot: &GitHubCopilotIntegration{Enabled: false},
						Windsurf:      &WindsurfIntegration{Enabled: false}, // zero value
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			cfgPath := filepath.Join(tmp, tc.name+".yaml")
			os.WriteFile(cfgPath, []byte(tc.yamlBody), 0644)

			loader := newYAMLLoader()
			cfg, err := loader.Load(cfgPath)
			require.NoError(t, err)

			if res := cmp.Diff(tc.expected, cfg); res != "" {
				t.Errorf("mismatch (-want +got):\n%s", res)
			}
		})
	}
}

func TestYamlLoader_Save(t *testing.T) {
	tmp := t.TempDir()

	loader := newYAMLLoader()

	tcs := []struct {
		name     string
		cfg      *Config
		expected string
	}{
		{
			name: "Full config",
			cfg: &Config{
				Settings: &Settings{
					CacheDir:     "/tmp/ajisai_cache",
					Experimental: true,
					Namespace:    "my_namespace",
				},
				Package: &Package{
					Name: "my_package",
					Exports: map[string]ExportedPresetDefinition{
						"preset1": {Prompts: []string{"prompts/prompt1.md"}, Rules: []string{"rules/rule1.json"}},
					},
				},
				Workspace: &Workspace{
					Imports: map[string]ImportedPackage{
						"import1": {Type: ImportTypeLocal, Details: LocalImportDetails{Path: "path/to/import1"}},
					},
					Integrations: &AgentIntegrations{
						Cursor:        &CursorIntegration{Enabled: true},
						GitHubCopilot: &GitHubCopilotIntegration{Enabled: false},
						Windsurf:      &WindsurfIntegration{Enabled: true},
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
			cfg: &Config{
				Settings: &Settings{
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
			cfg: &Config{
				Package: &Package{Name: "my_package"},
			},
			expected: `package:
  name: my_package
`,
		},
		{
			name: "Partial Workspace: imports",
			cfg: &Config{
				Workspace: &Workspace{
					Imports: map[string]ImportedPackage{
						"import1": {Type: ImportTypeLocal, Details: LocalImportDetails{Path: "path/to/import1"}},
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
			cfg: &Config{
				Workspace: &Workspace{
					Integrations: &AgentIntegrations{
						Cursor:        &CursorIntegration{Enabled: true},
						GitHubCopilot: &GitHubCopilotIntegration{Enabled: false},
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
