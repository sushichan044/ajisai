package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

// LocalFetcher implements the domain.ContentFetcher interface for local directory sources.
type LocalFetcher struct{}

// NewLocalFetcher creates a new LocalFetcher.
func NewLocalFetcher() *LocalFetcher {
	return &LocalFetcher{}
}

// Fetch copies content from the source local directory (defined in source.Details)
// to the destinationDir.
// It expects source.Details to be of type domain.LocalInputSourceDetails.
func (f *LocalFetcher) Fetch(ctx context.Context, source domain.InputSource, destinationDir string) error {
	// 1. Validate input source type and get details
	details, ok := source.Details.(domain.LocalInputSourceDetails)
	if !ok {
		return fmt.Errorf("LocalFetcher received unexpected source details type: %T", source.Details)
	}
	sourceDir := details.Path
	if sourceDir == "" {
		return errors.New("source path is empty in LocalInputSourceDetails")
	}

	// 2. Check if source directory exists and is a directory
	srcInfo, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source directory '%s' does not exist: %w", sourceDir, err)
		}
		return fmt.Errorf("failed to stat source directory '%s': %w", sourceDir, err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source path '%s' is not a directory", sourceDir)
	}

	// 3. Handle Destination Directory
	destInfo, err := os.Stat(destinationDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Destination does not exist, create it
			if mkErr := os.MkdirAll(destinationDir, 0755); mkErr != nil {
				return fmt.Errorf("failed to create destination directory '%s': %w", destinationDir, mkErr)
			}
		} else {
			// Other error stat-ing destination directory
			return fmt.Errorf("failed to stat destination directory '%s': %w", destinationDir, err)
		}
	} else {
		// Destination exists
		if !destInfo.IsDir() {
			return fmt.Errorf("destination path '%s' exists but is not a directory", destinationDir)
		}
		// Destination is a directory, clear its contents
		slog.Debug("Clearing existing destination directory contents", "path", destinationDir)
		dEntries, readErr := os.ReadDir(destinationDir)
		if readErr != nil {
			return fmt.Errorf("failed to read destination directory '%s' for clearing: %w", destinationDir, readErr)
		}
		for _, dEntry := range dEntries {
			pathToRemove := filepath.Join(destinationDir, dEntry.Name())
			if rmErr := os.RemoveAll(pathToRemove); rmErr != nil {
				return fmt.Errorf("failed to remove item '%s' in destination directory: %w", pathToRemove, rmErr)
			}
		}
	}

	// 4. Copy Content from Source to Destination
	slog.Debug("Starting content copy", "from", sourceDir, "to", destinationDir)
	err = filepath.WalkDir(sourceDir, func(srcPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("error walking source directory at '%s': %w", srcPath, walkErr)
		}

		// Calculate the relative path from the source base
		relPath, err := filepath.Rel(sourceDir, srcPath)
		if err != nil {
			// This should technically not happen if srcPath is within sourceDir
			return fmt.Errorf("failed to get relative path for '%s' base '%s': %w", srcPath, sourceDir, err)
		}

		// Calculate the corresponding destination path
		destPath := filepath.Join(destinationDir, relPath)

		if d.IsDir() {
			// Create the directory in the destination
			slog.Debug("Creating directory", "path", destPath)
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", destPath, err)
			}
		} else {
			// Copy the file
			slog.Debug("Copying file", "from", srcPath, "to", destPath)
			if err := copyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy file from '%s' to '%s': %w", srcPath, destPath, err)
			}
			// TODO: Optionally preserve permissions? os.Chmod(destPath, d.Type())? For now, default perms are used.
		}
		return nil
	})

	if err != nil {
		// If WalkDir failed, return the error
		return err
	}

	slog.Debug("Content copy completed successfully")
	return nil // Fetch completed successfully
}

// copyFile copies a single file from src to dest.
func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file '%s': %w", src, err)
	}
	defer sourceFile.Close()

	// Get source file info for permissions (optional)
	// srcInfo, err := sourceFile.Stat()
	// if err != nil {
	// 	 return fmt.Errorf("failed to stat source file '%s': %w", src, err)
	// }

	destFile, err := os.Create(dest) // Creates or truncates
	if err != nil {
		return fmt.Errorf("failed to create destination file '%s': %w", dest, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Optional: Set permissions
	// err = destFile.Chmod(srcInfo.Mode())
	// if err != nil {
	// 	 return fmt.Errorf("failed to set permissions on destination file '%s': %w", dest, err)
	// }

	return nil
}

// Compile-time check to ensure LocalFetcher implements ContentFetcher.
var _ domain.ContentFetcher = (*LocalFetcher)(nil)
