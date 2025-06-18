package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// Scheme defines the custom URI scheme for ajisai resources.
	Scheme = "ajisai"
)

// URI represents a unique identifier for a resolved resource.
// This represents logical structure rather than physical file paths.
type URI struct {
	Scheme  string     // "ajisai"
	Package string     // Package name (e.g., "local_rules")
	Preset  string     // Preset name (e.g., "default")
	Type    PresetType // Resource type ("rules" or "prompts")
	Path    string     // Hierarchical path within preset (e.g., "go-style/project")
}

// String converts the URI structure to its string representation.
// Example: "ajisai://local_rules/default/rules/go-style/project"
func (u *URI) String() string {
	return fmt.Sprintf("%s://%s/%s/%s/%s", u.Scheme, u.Package, u.Preset, u.Type, u.Path)
}

// GetInternalPath generates the relative path for writing to agent filesystem
// from the URI (e.g., "test-package/test-preset/my/rule.md").
func (u *URI) GetInternalPath(extension string) string {
	// u.Path is in format like "my/rule", so we add the extension
	return filepath.Join(u.Package, u.Preset, u.Path+extension)
}

// GetPathFromBaseDir calculates the relative file path from a base directory
// and returns the URI Path part (extension stripped).
// This consolidates the logic from utils.GetSlugFromBaseDir.
func GetPathFromBaseDir(baseAbsDir, targetPath string) (string, error) {
	cleanBaseAbsDir := filepath.Clean(baseAbsDir)
	cleanTargetPath := filepath.Clean(targetPath)

	if !filepath.IsAbs(cleanBaseAbsDir) || filepath.Ext(cleanBaseAbsDir) != "" {
		return "", fmt.Errorf("baseAbsDir %s is not an absolute directory", baseAbsDir)
	}

	if !filepath.IsAbs(cleanTargetPath) || filepath.Ext(cleanTargetPath) == "" {
		return "", fmt.Errorf("targetPath %s is not an absolute file", targetPath)
	}

	relPath, err := filepath.Rel(cleanBaseAbsDir, cleanTargetPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	if strings.HasPrefix(relPath, ".") {
		return "", fmt.Errorf("target path %s is not under base path %s", targetPath, baseAbsDir)
	}

	slashed := filepath.ToSlash(relPath)
	return strings.TrimSuffix(slashed, filepath.Ext(slashed)), nil
}
