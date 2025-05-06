package fetcher

import (
	"errors"
	"fmt"
	"os"

	cp "github.com/otiai10/copy"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

type LocalFetcherImpl struct{}

func LocalFetcher() *LocalFetcherImpl {
	return &LocalFetcherImpl{}
}

// Fetch copies content from the source local directory (defined in source.Details)
// to the destinationDir.
// It expects source.Details to be of type domain.LocalInputSourceDetails.
func (f *LocalFetcherImpl) Fetch(source domain.InputSource, destinationDir string) error {
	localDetails, ok := domain.GetInputSourceDetails[domain.LocalInputSourceDetails](source)
	if !ok {
		return &InvalidSourceTypeError{
			expectedType: "local",
			actualType:   source.Type,
			err:          fmt.Errorf("cannot fetch from source type: %s", source.Type),
		}
	}

	srcAbsDir, err := utils.ResolveAbsPath(localDetails.Path)
	if err != nil {
		return err
	}
	destAbsDir, err := utils.ResolveAbsPath(destinationDir)
	if err != nil {
		return err
	}

	// --- Destination Check ---
	destInfo, destStatErr := os.Stat(destAbsDir)
	if destStatErr != nil && !os.IsNotExist(destStatErr) {
		return fmt.Errorf("failed to stat destination directory '%s': %w", destAbsDir, destStatErr)
	}
	// Check if destination exists and is a file BEFORE checking source
	if destStatErr == nil && !destInfo.IsDir() {
		return fmt.Errorf("destination path '%s' exists but is not a directory", destAbsDir)
	}

	// --- Source Check ---
	if err := isValidSource(srcAbsDir); err != nil {
		return err
	}

	// --- Prepare Destination ---
	// Clean destination directory only if it existed as a directory before
	if destStatErr == nil && destInfo.IsDir() {
		if err := utils.EmptyDir(destAbsDir); err != nil {
			return fmt.Errorf("failed to empty destination directory '%s': %w", destAbsDir, err)
		}
	}
	// Ensure directory exists (might have been deleted by EmptyDir or never existed)
	if err := utils.EnsureDir(destAbsDir); err != nil {
		return fmt.Errorf("failed to ensure destination directory '%s': %w", destAbsDir, err)
	}

	// --- Copy ---
	if err := cp.Copy(srcAbsDir, destAbsDir); err != nil {
		return fmt.Errorf("failed to copy from '%s' to '%s': %w", srcAbsDir, destAbsDir, err)
	}

	return nil
}

// interface satisfaction check
var _ domain.ContentFetcher = (*LocalFetcherImpl)(nil)

// TODO: add directory structure check
func isValidSource(sourceDir string) error {
	if sourceDir == "" {
		return errors.New("source path cannot be empty")
	}

	exists, err := utils.IsDirExists(sourceDir)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("source directory '%s' does not exist", sourceDir)
	}

	return nil
}
