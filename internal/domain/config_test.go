package domain_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/domain"
)

func TestConfig_GetPresetRootInCache(t *testing.T) {
	tmpCacheDir := t.TempDir()

	tests := []struct {
		name           string
		config         *domain.Config
		presetName     string
		expectedPath   string
		expectedErrMsg string
	}{
		{
			name: "local input source",
			config: &domain.Config{
				Settings: domain.Settings{
					CacheDir: tmpCacheDir,
				},
				Inputs: map[string]domain.InputSource{
					"local_rules": {
						Type: domain.PresetSourceTypeLocal,
						Details: domain.LocalInputSourceDetails{
							Path: "./rules",
						},
					},
				},
			},
			presetName:     "local_rules",
			expectedPath:   filepath.Join(tmpCacheDir, "local_rules"),
			expectedErrMsg: "",
		},
		{
			name: "git input source without directory",
			config: &domain.Config{
				Settings: domain.Settings{
					CacheDir: tmpCacheDir,
				},
				Inputs: map[string]domain.InputSource{
					"git_rules": {
						Type: domain.PresetSourceTypeGit,
						Details: domain.GitInputSourceDetails{
							Repository: "https://github.com/example/repo",
							Revision:   "main",
						},
					},
				},
			},
			presetName:     "git_rules",
			expectedPath:   filepath.Join(tmpCacheDir, "git_rules"),
			expectedErrMsg: "",
		},
		{
			name: "git input source with directory",
			config: &domain.Config{
				Settings: domain.Settings{
					CacheDir: tmpCacheDir,
				},
				Inputs: map[string]domain.InputSource{
					"git_rules_with_dir": {
						Type: domain.PresetSourceTypeGit,
						Details: domain.GitInputSourceDetails{
							Repository: "https://github.com/example/repo",
							Revision:   "main",
							Directory:  "presets",
						},
					},
				},
			},
			presetName:     "git_rules_with_dir",
			expectedPath:   filepath.Join(tmpCacheDir, "git_rules_with_dir", "presets"),
			expectedErrMsg: "",
		},
		{
			name: "preset not found",
			config: &domain.Config{
				Settings: domain.Settings{
					CacheDir: tmpCacheDir,
				},
				Inputs: map[string]domain.InputSource{},
			},
			presetName:     "nonexistent",
			expectedPath:   "",
			expectedErrMsg: "preset nonexistent not found",
		},
		{
			name: "unsupported input source type",
			config: &domain.Config{
				Settings: domain.Settings{
					CacheDir: tmpCacheDir,
				},
				Inputs: map[string]domain.InputSource{
					"invalid": {
						Type: "invalid",
					},
				},
			},
			presetName:     "invalid",
			expectedPath:   "",
			expectedErrMsg: "unsupported input source type: invalid",
		},
		{
			name: "invalid local input source details",
			config: &domain.Config{
				Settings: domain.Settings{
					CacheDir: tmpCacheDir,
				},
				Inputs: map[string]domain.InputSource{
					"invalid_local": {
						Type:    domain.PresetSourceTypeLocal,
						Details: domain.GitInputSourceDetails{}, // 不正な詳細タイプ
					},
				},
			},
			presetName:     "invalid_local",
			expectedPath:   "",
			expectedErrMsg: fmt.Sprintf("invalid input source type: %s", domain.PresetSourceTypeLocal),
		},
		{
			name: "invalid git input source details",
			config: &domain.Config{
				Settings: domain.Settings{
					CacheDir: tmpCacheDir,
				},
				Inputs: map[string]domain.InputSource{
					"invalid_git": {
						Type:    domain.PresetSourceTypeGit,
						Details: domain.LocalInputSourceDetails{}, // 不正な詳細タイプ
					},
				},
			},
			presetName:     "invalid_git",
			expectedPath:   "",
			expectedErrMsg: fmt.Sprintf("invalid input source type: %s", domain.PresetSourceTypeGit),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := tt.config.GetPresetRootInCache(tt.presetName)
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
