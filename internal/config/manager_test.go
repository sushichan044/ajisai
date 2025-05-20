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

func TestManager_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	// Always fill all properties in the config.
	// Checking save-and-load is enough to check if the config is saved correctly.
	t.Run("successfully loads full-filled JSON config", func(t *testing.T) {
		// Create a test config file
		configPath := filepath.Join(tempDir, "config.json")
		testConfig := &config.Config{
			Settings: &config.Settings{
				CacheDir:     "/custom/cache/dir",
				Experimental: true,
				Namespace:    "test-namespace",
			},
			Workspace: &config.Workspace{
				Imports: map[string]config.ImportedPackage{
					"test-import": {
						Type: "git",
						Include: []string{
							"test-export",
							"typescript-react",
						},
						Details: config.GitImportDetails{
							Repository: "https://github.com/sushichan044/ajisai.git",
						},
					},
				},
				Integrations: &config.AgentIntegrations{
					Cursor: &config.CursorIntegration{
						Enabled: true,
					},
					GitHubCopilot: &config.GitHubCopilotIntegration{
						Enabled: true,
					},
					Windsurf: &config.WindsurfIntegration{
						Enabled: true,
					},
				},
			},
			Package: &config.Package{
				Name: "test-package",
				Exports: map[string]config.ExportedPresetDefinition{
					"go-guide": {
						Prompts: []string{"go-guide/prompts/**/*.md"},
						Rules:   []string{"go-guide/rules/**/*.md"},
					},
				},
			},
		}

		manager := config.New()

		if writeErr := manager.Save(configPath, testConfig); writeErr != nil {
			t.Fatalf("Failed to save config: %v", writeErr)
		}

		// Try to load the config
		loadedConfig, err := manager.Load(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		assert.True(t, cmp.Equal(loadedConfig, testConfig))
	})

	t.Run("fails with unsupported extension", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "config.unsupported")
		if writeErr := os.WriteFile(configPath, []byte("{}"), 0600); writeErr != nil {
			t.Fatalf("Failed to write test config: %v", writeErr)
		}

		manager := config.New()
		_, err := manager.Load(configPath)
		if err == nil {
			t.Error("Expected error for unsupported extension, but got nil")
		}
	})

	t.Run("fails with non-existent file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "non-existent.json")

		manager := config.New()
		_, err := manager.Load(configPath)
		if err == nil {
			t.Error("Expected error for non-existent file, but got nil")
		}
	})
}

func TestManagerSave(t *testing.T) {
	tempDir := t.TempDir()
	t.Run("fails with unsupported extension", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "config.unsupported")
		testConfig := &config.Config{}

		manager := config.New()
		err := manager.Save(configPath, testConfig)
		if err == nil {
			t.Error("Expected error for unsupported extension, but got nil")
		}
	})
}

//gocognit:ignore
func TestManagerApplyDefaults(t *testing.T) {
	t.Run("applies defaults to nil config", func(t *testing.T) {
		manager := config.New()
		cfg, err := manager.ApplyDefaults(nil)
		require.NoError(t, err)

		// Check settings defaults
		if cfg.Settings == nil {
			t.Fatal("Expected Settings to be initialized, but it's nil")
		}
		if cfg.Settings.CacheDir != "./.cache/ajisai" {
			t.Errorf("Expected default CacheDir %q, but got %q", "./.cache/ajisai", cfg.Settings.CacheDir)
		}
		if cfg.Settings.Namespace != "ajisai" {
			t.Errorf("Expected default Namespace %q, but got %q", "ajisai", cfg.Settings.Namespace)
		}

		// Check package defaults
		if cfg.Package == nil {
			t.Fatal("Expected Package to be initialized, but it's nil")
		}
		if cfg.Package.Exports == nil {
			t.Fatal("Expected Exports to be initialized, but it's nil")
		}

		// Check workspace defaults
		if cfg.Workspace == nil {
			t.Fatal("Expected Workspace to be initialized, but it's nil")
		}
		if cfg.Workspace.Imports == nil {
			t.Fatal("Expected Imports to be initialized, but it's nil")
		}
		if cfg.Workspace.Integrations == nil {
			t.Fatal("Expected Integrations to be initialized, but it's nil")
		}
	})

	t.Run("preserves existing values while applying defaults", func(t *testing.T) {
		inputConfig := &config.Config{
			Settings: &config.Settings{
				CacheDir:     "/custom/cache",
				Experimental: true,
				// Namespace left empty to test default
			},
			// Package and Workspace left nil to test default initialization
		}

		manager := config.New()
		cfg, err := manager.ApplyDefaults(inputConfig)
		require.NoError(t, err)

		// Verify original values are preserved
		if cfg.Settings.CacheDir != "/custom/cache" {
			t.Errorf("Expected CacheDir to be preserved as %q, but got %q", "/custom/cache", cfg.Settings.CacheDir)
		}
		if !cfg.Settings.Experimental {
			t.Error("Expected Experimental to be preserved as true, but it's false")
		}

		// Verify defaults are applied
		if cfg.Settings.Namespace != "ajisai" {
			t.Errorf("Expected default Namespace %q, but got %q", "ajisai", cfg.Settings.Namespace)
		}
		if cfg.Package == nil {
			t.Error("Expected Package to be initialized, but it's nil")
		}
		if cfg.Workspace == nil {
			t.Error("Expected Workspace to be initialized, but it's nil")
		}
	})
}

func TestManagerGetDefaultConfig(t *testing.T) {
	manager := config.New()
	cfg := manager.GetDefaultConfig()

	// Check settings defaults
	if cfg.Settings == nil {
		t.Fatal("Expected Settings to be initialized, but it's nil")
	}
	if cfg.Settings.CacheDir != "./.cache/ajisai" {
		t.Errorf("Expected default CacheDir %q, but got %q", "./.cache/ajisai", cfg.Settings.CacheDir)
	}
	if cfg.Settings.Namespace != "ajisai" {
		t.Errorf("Expected default Namespace %q, but got %q", "ajisai", cfg.Settings.Namespace)
	}

	// Check package defaults
	if cfg.Package == nil {
		t.Fatal("Expected Package to be initialized, but it's nil")
	}
	if cfg.Package.Exports == nil {
		t.Fatal("Expected Exports to be initialized, but it's nil")
	}

	// Check workspace defaults
	if cfg.Workspace == nil {
		t.Fatal("Expected Workspace to be initialized, but it's nil")
	}
	if cfg.Workspace.Imports == nil {
		t.Fatal("Expected Imports to be initialized, but it's nil")
	}
	if cfg.Workspace.Integrations == nil {
		t.Fatal("Expected Integrations to be initialized, but it's nil")
	}
}
