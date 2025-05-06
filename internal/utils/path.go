// Package utils provides generic utility functions for the application.
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveAbsPath converts the given path to an absolute path.
// If the path is already absolute, it is returned as is.
// If the path starts with "~", it's expanded to the user's home directory.
// If the path is empty, the current working directory is returned.
// Otherwise, the path is interpreted as relative to the current working directory.
func ResolveAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Resolve home directory expansion
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}

	// Handle empty path or relative path
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	if path == "" {
		return cwd, nil
	}

	return filepath.Join(cwd, path), nil
}
