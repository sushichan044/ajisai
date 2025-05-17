package config_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/config"
)

func TestConfig_GetImportedPackageCacheRoot(t *testing.T) {
	tmpCacheDir := t.TempDir()

	tests := []struct {
		name           string
		config         *config.Config
		presetName     string
		expectedPath   string
		expectedErrMsg string
	}{
		{
			name: "local_rules",
			config: &config.Config{
				Workspace: &config.Workspace{
					Imports: map[string]config.ImportedPackage{
						"local_rules": {
							Type: config.ImportTypeLocal,
							Details: config.LocalImportDetails{
								Path: "local_rules",
							},
						},
					},
				},
				Settings: &config.Settings{
					CacheDir: tmpCacheDir,
				},
			},
			presetName:     "local_rules",
			expectedPath:   filepath.Join(tmpCacheDir, "local_rules"),
			expectedErrMsg: "",
		},
		{
			name: "git_rules",
			config: &config.Config{
				Workspace: &config.Workspace{
					Imports: map[string]config.ImportedPackage{
						"git_rules": {
							Type: config.ImportTypeGit,
							Details: config.GitImportDetails{
								Repository: "https://github.com/sushichan044/ajisai-rules.git",
							},
						},
					},
				},
				Settings: &config.Settings{
					CacheDir: tmpCacheDir,
				},
			},
			presetName:     "git_rules",
			expectedPath:   filepath.Join(tmpCacheDir, "git_rules"),
			expectedErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := tt.config.GetImportedPackageCacheRoot(tt.presetName)
			if tt.expectedErrMsg != "" {
				require.EqualError(t, err, tt.expectedErrMsg)
				assert.Empty(t, path)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPath, path)
			}
		})
	}
}
