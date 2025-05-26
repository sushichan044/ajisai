package engine_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/config"
	"github.com/sushichan044/ajisai/internal/engine"
)

func TestNewEngine(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Settings: &config.Settings{},
		Workspace: &config.Workspace{
			Integrations: &config.AgentIntegrations{
				Cursor: &config.CursorIntegration{
					Enabled: false,
				},
				GitHubCopilot: &config.GitHubCopilotIntegration{
					Enabled: false,
				},
				Windsurf: &config.WindsurfIntegration{
					Enabled: false,
				},
			},
		},
	}

	// Execute
	engine, err := engine.NewEngine(cfg)

	// Verify
	require.NoError(t, err, "NewEngine should create engine without error")
	assert.NotNil(t, engine, "Created engine should not be nil")
}

func TestNewEngine_NilConfig(t *testing.T) {
	// Setup - nil config

	// Execute
	engine, err := engine.NewEngine(nil)

	// Verify
	require.Error(t, err, "NewEngine should error with nil config")
	assert.Nil(t, engine, "Engine should be nil on error")
	assert.Contains(t, err.Error(), "config is nil", "Error message should indicate config is nil")
}

func TestEngine_CleanCache_NonExistentDir(t *testing.T) {
	// Setup
	tempDir := t.TempDir()

	// Non-existent cache directory
	nonExistentDir := filepath.Join(tempDir, "non-existent")

	cfg := &config.Config{
		Settings: &config.Settings{
			CacheDir: nonExistentDir,
		},
		Workspace: &config.Workspace{
			Integrations: &config.AgentIntegrations{},
		},
	}

	engine, err := engine.NewEngine(cfg)
	require.NoError(t, err, "NewEngine should succeed with valid config")

	// Execute
	err = engine.CleanCache(false)

	// Verify
	assert.NoError(t, err, "CleanCache should not error when cache directory does not exist")
}

func TestEngine_CleanCache_ForceClean(t *testing.T) {
	// Setup
	tempDir := t.TempDir()

	// Create test file in cache dir
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err, "WriteFile should create test file successfully")

	cfg := &config.Config{
		Settings: &config.Settings{
			CacheDir: tempDir,
		},
		Workspace: &config.Workspace{
			Integrations: &config.AgentIntegrations{},
		},
	}

	engine, err := engine.NewEngine(cfg)
	require.NoError(t, err, "NewEngine should succeed with valid config")

	// Execute
	err = engine.CleanCache(true)

	// Verify
	require.NoError(t, err, "CleanCache with force should not error")
	_, err = os.Stat(testFile)
	require.True(t, os.IsNotExist(err), "Test file should not exist after force clean")
	_, err = os.Stat(tempDir)
	require.NoError(t, err, "Cache directory should still exist after force clean")
}

func TestEngine_CleanCache_SelectiveClean(t *testing.T) {
	// Setup
	tempDir := t.TempDir()

	// Create dirs for imported and not imported packages
	importedDir := filepath.Join(tempDir, "imported-pkg")
	notImportedDir := filepath.Join(tempDir, "not-imported-pkg")
	err := os.MkdirAll(importedDir, 0755)
	require.NoError(t, err, "MkdirAll should create imported directory")
	err = os.MkdirAll(notImportedDir, 0755)
	require.NoError(t, err, "MkdirAll should create not imported directory")

	// Create test files
	importedFile := filepath.Join(importedDir, "test.txt")
	notImportedFile := filepath.Join(notImportedDir, "test.txt")
	err = os.WriteFile(importedFile, []byte("test"), 0644)
	require.NoError(t, err, "WriteFile should create imported test file")
	err = os.WriteFile(notImportedFile, []byte("test"), 0644)
	require.NoError(t, err, "WriteFile should create not imported test file")

	cfg := &config.Config{
		Settings: &config.Settings{
			CacheDir: tempDir,
		},
		Workspace: &config.Workspace{
			Imports: map[string]config.ImportedPackage{
				"imported-pkg": {
					Type: "local",
				},
			},
			Integrations: &config.AgentIntegrations{},
		},
	}

	engine, err := engine.NewEngine(cfg)
	require.NoError(t, err, "NewEngine should succeed with valid config")

	// Execute
	err = engine.CleanCache(false)

	// Verify
	require.NoError(t, err, "CleanCache with selective clean should not error")
	_, err = os.Stat(importedFile)
	require.NoError(t, err, "Imported file should still exist after selective clean")
	_, err = os.Stat(notImportedFile)
	assert.True(t, os.IsNotExist(err), "Not imported file should not exist after selective clean")
}
