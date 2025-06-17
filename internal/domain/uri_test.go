package domain_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sushichan044/ajisai/internal/domain"
)

func TestURI_String(t *testing.T) {
	tests := []struct {
		name     string
		uri      domain.URI
		expected string
	}{
		{
			name: "basic rule URI",
			uri: domain.URI{
				Scheme:  domain.Scheme,
				Package: "local_rules",
				Preset:  "default",
				Type:    domain.RulesPresetType,
				Path:    "go-style/project",
			},
			expected: "ajisai://local_rules/default/rules/go-style/project",
		},
		{
			name: "basic prompt URI",
			uri: domain.URI{
				Scheme:  domain.Scheme,
				Package: "shared_prompts",
				Preset:  "dev",
				Type:    domain.PromptsPresetType,
				Path:    "debug/troubleshoot",
			},
			expected: "ajisai://shared_prompts/dev/prompts/debug/troubleshoot",
		},
		{
			name: "simple path without hierarchy",
			uri: domain.URI{
				Scheme:  domain.Scheme,
				Package: "test_pkg",
				Preset:  "test_preset",
				Type:    domain.RulesPresetType,
				Path:    "simple",
			},
			expected: "ajisai://test_pkg/test_preset/rules/simple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.uri.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestURI_GetInternalPath(t *testing.T) {
	tests := []struct {
		name      string
		uri       domain.URI
		extension string
		expected  string
	}{
		{
			name: "rule with hierarchy",
			uri: domain.URI{
				Package: "test-package",
				Preset:  "test-preset",
				Path:    "my/rule",
			},
			extension: ".md",
			expected:  filepath.Join("test-package", "test-preset", "my", "rule.md"),
		},
		{
			name: "prompt with simple path",
			uri: domain.URI{
				Package: "prompts",
				Preset:  "default",
				Path:    "simple",
			},
			extension: ".prompt.md",
			expected:  filepath.Join("prompts", "default", "simple.prompt.md"),
		},
		{
			name: "deeply nested path",
			uri: domain.URI{
				Package: "complex",
				Preset:  "nested",
				Path:    "deep/very/nested/path",
			},
			extension: ".mdc",
			expected:  filepath.Join("complex", "nested", "deep", "very", "nested", "path.mdc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.uri.GetInternalPath(tt.extension)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPathFromBaseDir(t *testing.T) {
	tests := []struct {
		name        string
		baseAbsDir  string
		targetPath  string
		expected    string
		expectError bool
	}{
		{
			name:       "simple file in base directory",
			baseAbsDir: "/base",
			targetPath: "/base/file.md",
			expected:   "file",
		},
		{
			name:       "file in subdirectory",
			baseAbsDir: "/base",
			targetPath: "/base/foo/bar.md",
			expected:   "foo/bar",
		},
		{
			name:       "deeply nested file",
			baseAbsDir: "/base",
			targetPath: "/base/foo/bar/baz.tar.gz",
			expected:   "foo/bar/baz.tar",
		},
		{
			name:       "file with complex extension",
			baseAbsDir: "/projects/ajisai",
			targetPath: "/projects/ajisai/rules/go/style.instructions.md",
			expected:   "rules/go/style.instructions",
		},
		{
			name:        "relative base directory",
			baseAbsDir:  "relative/path",
			targetPath:  "/absolute/file.md",
			expectError: true,
		},
		{
			name:        "base directory with extension",
			baseAbsDir:  "/base/file.txt",
			targetPath:  "/base/file.txt/other.md",
			expectError: true,
		},
		{
			name:        "relative target path",
			baseAbsDir:  "/base",
			targetPath:  "relative/file.md",
			expectError: true,
		},
		{
			name:        "target path without extension",
			baseAbsDir:  "/base",
			targetPath:  "/base/directory",
			expectError: true,
		},
		{
			name:        "target path outside base directory",
			baseAbsDir:  "/base",
			targetPath:  "/other/file.md",
			expectError: true,
		},
		{
			name:        "target path trying to escape with dots",
			baseAbsDir:  "/base",
			targetPath:  "/base/../other/file.md",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := domain.GetPathFromBaseDir(tt.baseAbsDir, tt.targetPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}