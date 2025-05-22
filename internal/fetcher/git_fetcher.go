package fetcher

import (
	"errors"
	"fmt"

	"github.com/sushichan044/ajisai/internal/config"
	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type GitFetcher struct {
	cmdRunner utils.CommandRunner
}

// NewGitFetcher creates a new GitFetcherImpl with the default command runner.
func NewGitFetcher() domain.PackageFetcher {
	return &GitFetcher{cmdRunner: &utils.DefaultCommandRunner{}}
}

// NewGitFetcherWithRunner creates a new GitFetcherImpl with a custom command runner (for testing).
func NewGitFetcherWithRunner(runner utils.CommandRunner) domain.PackageFetcher {
	return &GitFetcher{cmdRunner: runner}
}

// Fetch retrieves content from a Git repository.
// It clones the repo if destinationDir doesn't exist, otherwise updates it.
// If a specific revision is provided in the source, it checks out that revision.
// Otherwise, it pulls the latest changes from the default branch.
func (f *GitFetcher) Fetch(source config.ImportedPackage, destinationDir string) error {
	gitDetails, ok := config.GetImportDetails[config.GitImportDetails](source)
	if !ok {
		return &InvalidSourceTypeError{
			expectedType: config.ImportTypeGit,
			actualType:   source.Type,
			err:          fmt.Errorf("cannot fetch from source type: %s", source.Type),
		}
	}

	if gitDetails.Repository == "" {
		return errors.New("git repository URL cannot be empty")
	}

	destAbsDir, err := utils.ResolveAbsPath(destinationDir)
	if err != nil {
		return err
	}

	shouldPull, err := utils.IsDirExists(destAbsDir)
	if err != nil {
		return err
	}

	if !shouldPull {
		cmdArgs := []string{"clone", gitDetails.Repository, destAbsDir}
		return f.cmdRunner.Run("git", cmdArgs...)
	}

	// Reset hard to the latest commit
	if cleanErr := f.cmdRunner.RunInDir(destAbsDir, "git", "checkout", "."); cleanErr != nil {
		return fmt.Errorf("failed to clear dirty files in %s: %w", destAbsDir, cleanErr)
	}

	if gitDetails.Revision != "" {
		// Checkout specific revision
		fetchArgs := []string{"fetch", "origin"}
		if err = f.cmdRunner.RunInDir(destAbsDir, "git", fetchArgs...); err != nil {
			return fmt.Errorf(
				"failed to fetch updates for repository in %s: %w",
				destAbsDir,
				err,
			)
		}

		checkoutArgs := []string{"checkout", gitDetails.Revision}
		return f.cmdRunner.RunInDir(destAbsDir, "git", checkoutArgs...)
	}

	// Pull latest changes from default branch
	return f.cmdRunner.RunInDir(destAbsDir, "git", "pull")
}
