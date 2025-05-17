package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sushichan044/ajisai/internal/config"
)

func TestManagerLoad(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("successfully loads JSON config", func(t *testing.T) {
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
						Details: config.GitImportDetails{
							Repository: "https://github.com/sushichan044/ajisai.git",
						},
					},
				},
				Integrations: []config.AgentIntegration{
					{
						Target:  "cursor",
						Enabled: true,
					},
				},
			},
			Package: &config.Package{
				Name: "test-package",
				Exports: []config.ExportedPresetDefinition{
					{
						Name: "test-export",
					},
				},
			},
		}

		manager := config.NewManager()

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

		manager := config.NewManager()
		_, err := manager.Load(configPath)
		if err == nil {
			t.Error("Expected error for unsupported extension, but got nil")
		}
	})

	t.Run("fails with non-existent file", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "non-existent.json")

		manager := config.NewManager()
		_, err := manager.Load(configPath)
		if err == nil {
			t.Error("Expected error for non-existent file, but got nil")
		}
	})
}

func TestManagerSave(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("successfully saves JSON config", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "config.json")
		testConfig := &config.Config{
			Settings: &config.Settings{
				CacheDir:     "/custom/cache/dir",
				Experimental: true,
				Namespace:    "test-namespace",
			},
		}

		manager := config.NewManager()
		if err := manager.Save(configPath, testConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Read the saved file and verify its contents
		savedBytes, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read saved config: %v", err)
		}

		var savedConfig config.Config
		if jsonErr := json.Unmarshal(savedBytes, &savedConfig); jsonErr != nil {
			t.Fatalf("Failed to unmarshal saved config: %v", jsonErr)
		}

		if savedConfig.Settings.CacheDir != testConfig.Settings.CacheDir {
			t.Errorf(
				"Expected saved CacheDir %q, but got %q",
				testConfig.Settings.CacheDir,
				savedConfig.Settings.CacheDir,
			)
		}
		if savedConfig.Settings.Experimental != testConfig.Settings.Experimental {
			t.Errorf(
				"Expected saved Experimental %v, but got %v",
				testConfig.Settings.Experimental,
				savedConfig.Settings.Experimental,
			)
		}
		if savedConfig.Settings.Namespace != testConfig.Settings.Namespace {
			t.Errorf(
				"Expected saved Namespace %q, but got %q",
				testConfig.Settings.Namespace,
				savedConfig.Settings.Namespace,
			)
		}
	})

	t.Run("fails with unsupported extension", func(t *testing.T) {
		configPath := filepath.Join(tempDir, "config.unsupported")
		testConfig := &config.Config{}

		manager := config.NewManager()
		err := manager.Save(configPath, testConfig)
		if err == nil {
			t.Error("Expected error for unsupported extension, but got nil")
		}
	})
}

//gocognit:ignore
func TestManagerApplyDefaults(t *testing.T) {
	t.Run("applies defaults to nil config", func(t *testing.T) {
		manager := config.NewManager()
		cfg, err := manager.ApplyDefaults(nil)
		if err != nil {
			t.Fatalf("Failed to apply defaults: %v", err)
		}

		// Check settings defaults
		if cfg.Settings == nil {
			t.Fatal("Expected Settings to be initialized, but it's nil")
		}
		if cfg.Settings.CacheDir != "./.ajisai/cache" {
			t.Errorf("Expected default CacheDir %q, but got %q", "./.ajisai/cache", cfg.Settings.CacheDir)
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

		manager := config.NewManager()
		cfg, err := manager.ApplyDefaults(inputConfig)
		if err != nil {
			t.Fatalf("Failed to apply defaults: %v", err)
		}

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
	manager := config.NewManager()
	cfg := manager.GetDefaultConfig()

	// Check settings defaults
	if cfg.Settings == nil {
		t.Fatal("Expected Settings to be initialized, but it's nil")
	}
	if cfg.Settings.CacheDir != "./.ajisai/cache" {
		t.Errorf("Expected default CacheDir %q, but got %q", "./.ajisai/cache", cfg.Settings.CacheDir)
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
