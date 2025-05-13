package fetcher

import (
	"errors"
	"fmt"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type GitFetcher struct {
	cmdRunner utils.CommandRunner
}

// NewGitFetcher creates a new GitFetcherImpl with the default command runner.
func NewGitFetcher() domain.ContentFetcher {
	return &GitFetcher{cmdRunner: &utils.DefaultCommandRunner{}}
}

// NewGitFetcherWithRunner creates a new GitFetcherImpl with a custom command runner (for testing).
func NewGitFetcherWithRunner(runner utils.CommandRunner) domain.ContentFetcher {
	return &GitFetcher{cmdRunner: runner}
}

// Fetch retrieves content from a Git repository.
// It clones the repo if destinationDir doesn't exist, otherwise updates it.
// If a specific revision is provided in the source, it checks out that revision.
// Otherwise, it pulls the latest changes from the default branch.
func (f *GitFetcher) Fetch(source domain.InputSource, destinationDir string) error {
	gitSource, ok := domain.GetInputSourceDetails[domain.GitInputSourceDetails](source)
	if !ok {
		return &InvalidSourceTypeError{
			expectedType: domain.PresetSourceTypeGit,
			actualType:   source.Type,
			err:          fmt.Errorf("cannot fetch from source type: %s", source.Type),
		}
	}

	if gitSource.Repository == "" {
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
		cmdArgs := []string{"clone", gitSource.Repository, destAbsDir}
		return f.cmdRunner.Run("git", cmdArgs...)
	}

	if gitSource.Revision != "" {
		// Checkout specific revision
		fetchArgs := []string{"fetch", "origin"}
		if err = f.cmdRunner.RunInDir(destAbsDir, "git", fetchArgs...); err != nil {
			return fmt.Errorf(
				"failed to fetch updates for repository in %s: %w",
				destAbsDir,
				err,
			)
		}

		checkoutArgs := []string{"checkout", gitSource.Revision}
		return f.cmdRunner.RunInDir(destAbsDir, "git", checkoutArgs...)
	}

	// Pull latest changes from default branch
	pullArgs := []string{"pull", "origin"}
	return f.cmdRunner.RunInDir(destAbsDir, "git", pullArgs...)
}
