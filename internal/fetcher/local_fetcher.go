package fetcher

import (
	"errors"
	"fmt"

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
