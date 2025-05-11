package utils

import (
	"fmt"
	"os"
)

// EmptyDir removes all files and directories in the given path.
func EmptyDir(path string) error {
	absPath, err := ResolveAbsPath(path)
	if err != nil {
		return err
	}

	if err = os.RemoveAll(absPath); err != nil {
		return err
	}

	return nil
}

// EnsureDir creates a directory if it doesn't exist, and returns an error if it's not a directory.
func EnsureDir(path string) error {
	absPath, err := ResolveAbsPath(path)
	if err != nil {
		return err
	}

	stat, err := os.Stat(absPath)

	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(absPath, 0750)
		}
		return err
	}

	if !stat.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", absPath)
	}

	return nil
}

// IsDirExists checks if a path exists and is a directory.
// Returns an error if the path exists but is not a directory, or if os.Stat fails for other reasons.
func IsDirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Does not exist, not an error for this check
		}
		return false, err // Other stat error
	}
	// Exists, check if it is a directory
	if !info.IsDir() {
		return false, fmt.Errorf("path '%s' exists but is not a directory", path)
	}
	return true, nil // Exists and is a directory
}
