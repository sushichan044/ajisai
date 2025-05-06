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
