package utils

import (
	"fmt"
	"io"
	"os"
)

// AtomicWriteFile writes a file atomically.
// File permissions are set to 0600.
func AtomicWriteFile(path string, reader io.Reader) error {
	tmp, tmpErr := os.CreateTemp("", "ajisai-atomic-*.tmp")
	if tmpErr != nil {
		return fmt.Errorf("failed to create temporary file: %w", tmpErr)
	}

	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	defer tmp.Close()

	if _, err := io.Copy(tmp, reader); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err := os.Chmod(tmpName, 0o600); err != nil {
		return fmt.Errorf("failed to set permissions for temporary file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("failed to rename temporary file to target file %s: %w", path, err)
	}

	return nil
}
