package fetcher

import (
	"errors"
	"fmt"

	cp "github.com/otiai10/copy"

	"github.com/sushichan044/ajisai/internal/config"
	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type LocalFetcher struct{}

func NewLocalFetcher() domain.PackageFetcher {
	return &LocalFetcher{}
}

// Fetch copies content from the source local directory (defined in source.Details)
// to the destinationDir.
// It expects source.Details to be of type domain.LocalInputSourceDetails.
func (f *LocalFetcher) Fetch(source config.ImportedPackage, destinationDir string) error {
	localDetails, ok := config.GetImportDetails[config.LocalImportDetails](source)
	if !ok {
		return &InvalidSourceTypeError{
			expectedType: config.ImportTypeLocal,
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

	if err = isValidSource(srcAbsDir); err != nil {
		return err
	}

	// Clean destination directory
	if err = utils.EmptyDir(destAbsDir); err != nil {
		return err
	}
	if err = utils.EnsureDir(destAbsDir); err != nil {
		return err
	}

	if err = cp.Copy(srcAbsDir, destAbsDir); err != nil {
		return err
	}

	return nil
}

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
