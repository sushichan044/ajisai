package fetcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

// commandRunner defines the signature for running external commands, allowing mocking.
type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// GitFetcher fetches content from a Git repository using the git command.
// Ensure 'git' command is available in the system PATH.
type GitFetcher struct {
	Runner commandRunner
	// lookPathFunc is primarily used during construction, not usually needed during Fetch
	// lookPathFunc func(file string) (string, error)
}

// defaultCommandRunner executes a command and returns its combined output.
func defaultCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// NewGitFetcher creates a new instance of GitFetcher.
// It checks if the 'git' command is available in the system PATH.
func NewGitFetcher() (*GitFetcher, error) {
	_, err := exec.LookPath("git") // Still check for git existence
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("'git' command not found in PATH: %w", exec.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to check for 'git' command: %w", err)
	}
	return &GitFetcher{
		Runner: defaultCommandRunner,
	}, nil
}

// Compile-time check to ensure GitFetcher implements ContentFetcher.
var _ domain.ContentFetcher = (*GitFetcher)(nil)

// Fetch retrieves content from a Git repository.
// It clones the repo if destinationDir doesn't exist, otherwise updates it.
// If a specific revision is provided in the source, it checks out that revision.
// Otherwise, it pulls the latest changes from the default branch.
func (f *GitFetcher) Fetch(ctx context.Context, source domain.InputSource, destinationDir string) error {
	gitSource, ok := source.Details.(*domain.GitInputSourceDetails)
	if !ok {
		return fmt.Errorf("invalid input source type for GitFetcher: %T", source.Details)
	}

	if gitSource.Repository == "" {
		return errors.New("git repository URL is empty")
	}

	// Use os.Stat directly here, test can mock it via osStat stub if needed for specific scenarios
	// but primary mocking is via command runner
	_, err := os.Stat(destinationDir)
	var dirExists bool
	if err == nil {
		dirExists = true
	} else if errors.Is(err, os.ErrNotExist) {
		dirExists = false
	} else {
		// Other error from os.Stat
		return fmt.Errorf("failed to check destination directory %s: %w", destinationDir, err)
	}

	if !dirExists {
		// Initial clone
		fmt.Printf("Cloning repository %s into %s...\n", gitSource.Repository, destinationDir)
		cmdArgs := []string{"clone", gitSource.Repository, destinationDir}
		// Use the exported Runner field
		output, err := f.Runner(ctx, "git", cmdArgs...)
		if err != nil {
			return fmt.Errorf("failed to clone repository %s: %w\nOutput:\n%s", gitSource.Repository, err, string(output))
		}
		fmt.Printf("Clone successful.\n")
		// No need to checkout revision or pull after initial clone (git clone handles default branch)
		return nil
	}

	// Directory exists, handle update
	fmt.Printf("Updating repository in %s...\n", destinationDir)

	if gitSource.Revision != "" {
		// Checkout specific revision
		fmt.Printf("Fetching latest changes for revision %s...\n", gitSource.Revision)
		fetchArgs := []string{"-C", destinationDir, "fetch", "origin"}
		output, err := f.Runner(ctx, "git", fetchArgs...)
		if err != nil {
			return fmt.Errorf("failed to fetch updates for repository in %s: %w\nOutput:\n%s", destinationDir, err, string(output))
		}

		fmt.Printf("Checking out revision %s...\n", gitSource.Revision)
		checkoutArgs := []string{"-C", destinationDir, "checkout", gitSource.Revision}
		output, err = f.Runner(ctx, "git", checkoutArgs...)
		if err != nil {
			// Add specific error context for checkout failure
			return fmt.Errorf("failed to checkout revision %s in %s: %w\nOutput:\n%s", gitSource.Revision, destinationDir, err, string(output))
		}
		fmt.Printf("Successfully checked out revision %s.\n", gitSource.Revision)
		return nil
	} else {
		// Pull latest changes from default branch
		fmt.Printf("Pulling latest changes for default branch...\n")
		pullArgs := []string{"-C", destinationDir, "pull", "origin"} // Assuming origin is the default remote
		output, err := f.Runner(ctx, "git", pullArgs...)
		if err != nil {
			// If err is not nil, it's a real error
			return fmt.Errorf("failed to pull latest changes for repository in %s: %w\nOutput:\n%s", destinationDir, err, string(output))
		}
		fmt.Printf("Pull successful.\n%s", string(output)) // Include output for info
		return nil
	}

	// This part should not be reached
	// return fmt.Errorf("update logic not fully implemented yet") // This line was likely commented out already
}
