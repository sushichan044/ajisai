package utils_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/utils"
)

func TestGetSlugFromPath(t *testing.T) {
	tests := []struct {
		name string
		// baseAbsDir must be an absolute directory path.
		baseAbsDir string
		// targetPath must be an absolute file path.
		targetPath    string
		expectedSlug  string
		expectedError bool
	}{
		{
			name:          "valid path",
			baseAbsDir:    "/base",
			targetPath:    "/base/foo/bar.md",
			expectedSlug:  "foo/bar",
			expectedError: false,
		},
		{
			name:          "valid path with extension",
			baseAbsDir:    "/another/base",
			targetPath:    "/another/base/baz/qux.txt",
			expectedSlug:  "baz/qux",
			expectedError: false,
		},
		{
			name:          "valid path, no subdir",
			baseAbsDir:    "/base",
			targetPath:    "/base/my_prompt.md",
			expectedSlug:  "my_prompt",
			expectedError: false,
		},
		{
			name:          "targetPath is not under baseAbsDir",
			baseAbsDir:    "/base",
			targetPath:    "/other/foo/bar.md",
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "targetPath has no extension",
			baseAbsDir:    "/base",
			targetPath:    "/base/foo/bar", // No extension
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "targetPath is a hidden file",
			baseAbsDir:    "/base",
			targetPath:    "/base/foo/.bar.md", // Hidden file, slug will be "foo/.bar"
			expectedSlug:  "foo/.bar",
			expectedError: false,
		},
		{
			name:          "baseAbsDir and targetPath are same directory (invalid baseAbsDir)",
			baseAbsDir:    "/base/foo",
			targetPath:    "/base/foo", // targetPath is not a file
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "baseAbsDir is a file path (invalid)",
			baseAbsDir:    "/base/foo.md", // baseAbsDir is not a directory
			targetPath:    "/base/foo.md",
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "baseAbsDir is relative path (invalid)",
			baseAbsDir:    "base", // Relative path
			targetPath:    "/base/foo/bar.md",
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "targetPath is relative path (invalid)",
			baseAbsDir:    "/base",
			targetPath:    "base/foo/bar.md", // Relative path
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "baseAbsDir is not an absolute path",
			baseAbsDir:    "not/abs",
			targetPath:    "/not/abs/foo/bar.md",
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "targetPath is not an absolute path",
			baseAbsDir:    "/abs/base",
			targetPath:    "not/abs/file.md",
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "targetPath is a directory (invalid)",
			baseAbsDir:    "/base",
			targetPath:    "/base/foo/", // targetPath is a directory
			expectedSlug:  "",
			expectedError: true,
		},
		{
			name:          "targetPath directly under baseAbsDir, multi-dot extension",
			baseAbsDir:    "/base",
			targetPath:    "/base/archive.tar.gz",
			expectedSlug:  "archive.tar",
			expectedError: false,
		},
		{
			name:          "targetPath in subdirectory, multi-dot extension",
			baseAbsDir:    "/base",
			targetPath:    "/base/sub/archive.tar.gz",
			expectedSlug:  "sub/archive.tar",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to platform specific paths for testing
			// Test cases use Unix-style paths for clarity.
			// These are converted to platform-specific paths before calling the function.
			platformBaseAbsDir := filepath.FromSlash(tt.baseAbsDir)
			platformTargetPath := filepath.FromSlash(tt.targetPath)

			slug, err := utils.GetSlugFromBaseDir(platformBaseAbsDir, platformTargetPath)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSlug, slug)
			}
		})
	}
}
