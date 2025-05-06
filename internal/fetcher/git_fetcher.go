package fetcher

import (
	"errors"
	"fmt"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

type GitFetcherImpl struct {
	cmdRunner utils.CommandRunner
}

// GitFetcher creates a new GitFetcherImpl with the default command runner.
func GitFetcher() *GitFetcherImpl {
	return &GitFetcherImpl{cmdRunner: &utils.DefaultCommandRunner{}}
}

// GitFetcherWithRunner creates a new GitFetcherImpl with a custom command runner (for testing).
func GitFetcherWithRunner(runner utils.CommandRunner) *GitFetcherImpl {
	return &GitFetcherImpl{cmdRunner: runner}
}

// Compile-time check to ensure GitFetcher implements ContentFetcher.
var _ domain.ContentFetcher = (*GitFetcherImpl)(nil)

// Fetch retrieves content from a Git repository.
// It clones the repo if destinationDir doesn't exist, otherwise updates it.
// If a specific revision is provided in the source, it checks out that revision.
// Otherwise, it pulls the latest changes from the default branch.
func (f *GitFetcherImpl) Fetch(source domain.InputSource, destinationDir string) error {
	gitSource, ok := domain.GetInputSourceDetails[domain.GitInputSourceDetails](source)
	if !ok {
		return &InvalidSourceTypeError{
			expectedType: "git",
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
