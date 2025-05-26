package loader_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/config"
	"github.com/sushichan044/ajisai/internal/loader"
)

func TestNewAgentPresetPackageLoader(t *testing.T) {
	// Setup
	cfg := &config.Config{}

	// Execute
	loader := loader.NewAgentPresetPackageLoader(cfg)

	// Verify
	assert.NotNil(t, loader, "NewAgentPresetPackageLoader should return non-nil loader")
}

func TestLoadAgentPresetPackage_PackageNotImported(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Workspace: &config.Workspace{
			Imports: map[string]config.ImportedPackage{
				// Empty map - no imports
			},
		},
	}
	loader := loader.NewAgentPresetPackageLoader(cfg)

	// Execute
	pkg, err := loader.LoadAgentPresetPackage("non-existent-package")

	// Verify
	require.Error(t, err, "LoadAgentPresetPackage should return error for non-imported package")
	assert.Nil(t, pkg, "Package should be nil when not imported")
	assert.Contains(t, err.Error(), "not imported", "Error message should indicate package is not imported")
}

func TestResolvePackageManifest_DefaultFallback(t *testing.T) {
	// Setup - create a temp directory to simulate cache
	tempDir := t.TempDir()

	// Setup package directory
	packageName := "test-package"
	packageDir := filepath.Join(tempDir, packageName)
	err := os.MkdirAll(packageDir, 0755)
	require.NoError(t, err, "MkdirAll should create package directory")

	// Create a config that points to the temp dir
	cfg := &config.Config{
		Settings: &config.Settings{
			CacheDir: tempDir,
		},
		Workspace: &config.Workspace{
			Imports: map[string]config.ImportedPackage{
				packageName: {
					Type: "local",
				},
			},
		},
	}

	l := loader.NewAgentPresetPackageLoader(cfg)

	// Execute - we'll intentionally not create a manifest file to test the fallback
	manifest, err := l.ResolvePackageManifest(packageName)

	// Verify
	require.NoError(t, err, "ResolvePackageManifest should fallback to default preset without error")
	assert.Equal(t, packageName, manifest.Name, "Package name should match input")
	assert.Contains(t, manifest.Exports, config.DefaultPresetName, "Manifest should contain default preset")
	assert.Contains(
		t,
		manifest.Exports[config.DefaultPresetName].Prompts,
		"prompts/**/*.md",
		"Default preset should include prompts glob pattern",
	)
	assert.Contains(
		t,
		manifest.Exports[config.DefaultPresetName].Rules,
		"rules/**/*.md",
		"Default preset should include rules glob pattern",
	)
}

func TestLoadAgentPresetPackage_Success(t *testing.T) {
	// Setup - create a temp directory with valid structure
	tempDir := t.TempDir()

	// Setup package directory
	packageName := "test-package"
	packageDir := filepath.Join(tempDir, packageName)
	err := os.MkdirAll(packageDir, 0755)
	require.NoError(t, err, "MkdirAll should create package directory")

	// Create nested directories
	promptsDir := filepath.Join(packageDir, "prompts")
	rulesDir := filepath.Join(packageDir, "rules")
	err = os.MkdirAll(promptsDir, 0755)
	require.NoError(t, err, "MkdirAll should create prompts directory")
	err = os.MkdirAll(rulesDir, 0755)
	require.NoError(t, err, "MkdirAll should create rules directory")

	// Create valid markdown files
	promptContent := `---
description: "Test Prompt"
---
# Test Prompt
This is a test prompt.`

	ruleContent := `---
description: "Test Rule"
attach: "always"
---
# Test Rule
This is a test rule.`

	err = os.WriteFile(filepath.Join(promptsDir, "prompt.md"), []byte(promptContent), 0644)
	require.NoError(t, err, "WriteFile should create prompt file")
	err = os.WriteFile(filepath.Join(rulesDir, "rule.md"), []byte(ruleContent), 0644)
	require.NoError(t, err, "WriteFile should create rule file")

	// Create manifest file
	manifestContent := `package:
  exports:
    default:
      prompts: ["prompts/**/*.md"]
      rules: ["rules/**/*.md"]
`
	manifestPath := filepath.Join(packageDir, "ajisai.yml")
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	require.NoError(t, err, "WriteFile should create manifest file")

	// Setup config
	cfg := &config.Config{
		Settings: &config.Settings{
			CacheDir: tempDir,
		},
		Workspace: &config.Workspace{
			Imports: map[string]config.ImportedPackage{
				packageName: {
					Type: "local",
					Include: []string{
						"default",
					},
				},
			},
		},
	}

	loader := loader.NewAgentPresetPackageLoader(cfg)

	// Execute
	pkg, err := loader.LoadAgentPresetPackage(packageName)

	// Verify
	require.NoError(t, err, "LoadAgentPresetPackage should load package without error")
	assert.Equal(t, packageName, pkg.PackageName, "Package name should match input")
	assert.Len(t, pkg.Presets, 1, "Package should contain exactly one preset")
	assert.Equal(t, "default", pkg.Presets[0].Name, "Preset name should be 'default'")
	assert.Len(t, pkg.Presets[0].Prompts, 1, "Preset should contain exactly one prompt")
	assert.Len(t, pkg.Presets[0].Rules, 1, "Preset should contain exactly one rule")
}
