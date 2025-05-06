package fetcher

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	Logger *slog.Logger
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
func NewGitFetcher(logger *slog.Logger) (*GitFetcher, error) {
	_, err := exec.LookPath("git")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("'git' command not found in PATH: %w", exec.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to check for 'git' command: %w", err)
	}
	// Ensure logger is not nil, provide default if necessary
	useLogger := logger
	if useLogger == nil {
		useLogger = slog.Default() // Or slog.New(slog.DiscardHandler) if preferred
	}
	return &GitFetcher{
		Runner: defaultCommandRunner,
		Logger: useLogger,
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
		f.Logger.InfoContext(ctx, "Cloning repository", "url", gitSource.Repository, "dest", destinationDir)
		cmdArgs := []string{"clone", gitSource.Repository, destinationDir}
		// Use the exported Runner field
		output, err := f.Runner(ctx, "git", cmdArgs...)
		if err != nil {
			f.Logger.ErrorContext(
				ctx,
				"Failed to clone repository",
				"url",
				gitSource.Repository,
				"error",
				err,
				"output",
				string(output),
			)
			return fmt.Errorf(
				"failed to clone repository %s: %w\nOutput:\n%s",
				gitSource.Repository,
				err,
				string(output),
			)
		}
		f.Logger.InfoContext(ctx, "Clone successful", "dest", destinationDir)
		// No need to checkout revision or pull after initial clone (git clone handles default branch)
		return nil
	}

	// Directory exists, handle update
	f.Logger.InfoContext(ctx, "Updating repository", "dest", destinationDir)

	if gitSource.Revision != "" {
		// Checkout specific revision
		f.Logger.InfoContext(ctx, "Fetching latest changes", "dest", destinationDir, "revision", gitSource.Revision)
		fetchArgs := []string{"-C", destinationDir, "fetch", "origin"}
		output, err := f.Runner(ctx, "git", fetchArgs...)
		if err != nil {
			f.Logger.ErrorContext(
				ctx,
				"Failed to fetch updates",
				"dest",
				destinationDir,
				"error",
				err,
				"output",
				string(output),
			)
			return fmt.Errorf(
				"failed to fetch updates for repository in %s: %w\nOutput:\n%s",
				destinationDir,
				err,
				string(output),
			)
		}

		f.Logger.InfoContext(ctx, "Checking out revision", "dest", destinationDir, "revision", gitSource.Revision)
		checkoutArgs := []string{"-C", destinationDir, "checkout", gitSource.Revision}
		output, err = f.Runner(ctx, "git", checkoutArgs...)
		if err != nil {
			f.Logger.ErrorContext(
				ctx,
				"Failed to checkout revision",
				"dest",
				destinationDir,
				"revision",
				gitSource.Revision,
				"error",
				err,
				"output",
				string(output),
			)
			return fmt.Errorf(
				"failed to checkout revision %s in %s: %w\nOutput:\n%s",
				gitSource.Revision,
				destinationDir,
				err,
				string(output),
			)
		}
		f.Logger.InfoContext(
			ctx,
			"Successfully checked out revision",
			"dest",
			destinationDir,
			"revision",
			gitSource.Revision,
		)
		return nil
	} else {
		// Pull latest changes from default branch
		f.Logger.InfoContext(ctx, "Pulling latest changes for default branch", "dest", destinationDir)
		pullArgs := []string{"-C", destinationDir, "pull", "origin"} // Assuming origin is the default remote
		output, err := f.Runner(ctx, "git", pullArgs...)
		if err != nil {
			f.Logger.ErrorContext(ctx, "Failed to pull latest changes", "dest", destinationDir, "error", err, "output", string(output))
			return fmt.Errorf("failed to pull latest changes for repository in %s: %w\nOutput:\n%s", destinationDir, err, string(output))
		}
		f.Logger.InfoContext(ctx, "Pull successful", "dest", destinationDir)
		return nil
	}
}
