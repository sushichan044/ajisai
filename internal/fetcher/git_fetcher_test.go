package fetcher_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/fetcher"
)

func TestGitFetcher_ImplementsContentFetcher(t *testing.T) {
	// This test primarily checks at compile time if GitFetcher satisfies the interface.
	// An explicit check can also be added.
	var _ domain.ContentFetcher = (*fetcher.GitFetcher)(nil)

	// Optional: Instantiate and check
	gf, err := fetcher.NewGitFetcher(nil) // Pass nil logger for interface check
	require.NoError(t, err)               // Assuming NewGitFetcher handles nil logger
	assert.NotNil(t, gf, "GitFetcher instance should not be nil")
}

func TestGitFetcher_Fetch_InitialClone(t *testing.T) {
	ctx := t.Context()
	destDir := t.TempDir()
	require.NoError(t, os.RemoveAll(destDir)) // Ensure destDir does not exist initially

	repoURL := "https://github.com/example/repo.git"
	source := domain.InputSource{
		Type:    "git",
		Details: &domain.GitInputSourceDetails{Repository: repoURL},
	}

	// Arrange Mocks (os.Stat can still be mocked if needed, but let's focus on runner)
	// Simulate directory not existing by ensuring it's removed

	var executedCommand []string
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		require.Equal(t, "git", name)
		executedCommand = append([]string{name}, args...)
		// Simulate success
		return []byte("Cloning into '" + destDir + "'...\ndone."), nil
	}

	// Act
	fetcherInstance := fetcher.GitFetcher{
		Runner: mockRunner,
		Logger: slog.New(slog.DiscardHandler), // Use a discard logger for testing
	}
	err := fetcherInstance.Fetch(ctx, source, destDir)

	// Assert
	assert.NoError(t, err, "Fetch should succeed for initial clone")
	expectedCmd := []string{"git", "clone", repoURL, destDir}
	assert.Equal(t, expectedCmd, executedCommand, "Expected git clone command")
}

func TestGitFetcher_Fetch_InitialClone_Failure(t *testing.T) {
	// No need for resetMocks() or global stubs

	ctx := t.Context()
	destDir := t.TempDir()
	require.NoError(t, os.RemoveAll(destDir))

	repoURL := "invalid-url"
	source := domain.InputSource{
		Type:    "git",
		Details: &domain.GitInputSourceDetails{Repository: repoURL},
	}

	// Arrange Mocks
	var executedCommand []string
	mockError := errors.New("mock git execution error")
	mockOutput := []byte("fatal: repository not found")
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		require.Equal(t, "git", name)
		executedCommand = append([]string{name}, args...)
		// Simulate failure
		return mockOutput, mockError
	}

	// Act
	fetcherInstance := fetcher.GitFetcher{
		Runner: mockRunner,
		Logger: slog.New(slog.DiscardHandler),
	}
	err := fetcherInstance.Fetch(ctx, source, destDir)

	// Assert
	assert.Error(t, err, "Fetch should fail when git clone fails")
	assert.Contains(t, err.Error(), "failed to clone repository", "Error message should indicate clone failure")
	assert.ErrorIs(t, err, mockError, "Error should wrap the runner error")
	assert.Contains(t, err.Error(), string(mockOutput), "Error message should contain output from command")
	expectedCmd := []string{"git", "clone", repoURL, destDir}
	assert.Equal(t, expectedCmd, executedCommand, "Expected git clone command")
}

func TestGitFetcher_Fetch_CheckoutRevision(t *testing.T) {
	ctx := t.Context()
	destDir := t.TempDir() // Simulate existing directory
	revision := "v1.0.0"
	source := domain.InputSource{
		Type: "git",
		Details: &domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	// Arrange Mocks
	var executedCommands [][]string
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		require.Equal(t, "git", name)
		cmd := append([]string{name}, args...)
		executedCommands = append(executedCommands, cmd)

		// Find the actual git command (fetch or checkout) after -C option
		var gitCommand string
		for i, arg := range args {
			if arg == "-C" && i+1 < len(args) {
				// Skip the directory path
				i++
				continue
			}
			if arg == "fetch" || arg == "checkout" {
				gitCommand = arg
				break
			}
		}

		// Simulate success/failure based on the found command
		switch gitCommand {
		case "fetch":
			return []byte("Fetched successfully"), nil
		case "checkout":
			return []byte("Switched to revision '" + revision + "'"), nil
		default:
			return nil, fmt.Errorf("unexpected git command structure: %v", args)
		}
	}

	// Act
	fetcherInstance := fetcher.GitFetcher{
		Runner: mockRunner,
		Logger: slog.New(slog.DiscardHandler),
	}
	err := fetcherInstance.Fetch(ctx, source, destDir)

	// Assert
	assert.NoError(t, err, "Fetch should succeed when checking out revision")
	require.Len(t, executedCommands, 2, "Expected two git commands")
	assert.Equal(
		t,
		[]string{"git", "-C", destDir, "fetch", "origin"},
		executedCommands[0],
		"First command should be git fetch with -C",
	)
	assert.Equal(
		t,
		[]string{"git", "-C", destDir, "checkout", revision},
		executedCommands[1],
		"Second command should be git checkout with -C",
	)
}

func TestGitFetcher_Fetch_CheckoutRevision_Failure(t *testing.T) {
	ctx := t.Context()
	destDir := t.TempDir()
	revision := "invalid-revision"
	source := domain.InputSource{
		Type: "git",
		Details: &domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	// Arrange Mocks
	var executedCommands [][]string
	mockError := errors.New("mock git checkout error")
	mockOutput := []byte("error: pathspec 'invalid-revision' did not match any file(s) known to git")
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		require.Equal(t, "git", name)
		cmd := append([]string{name}, args...)
		executedCommands = append(executedCommands, cmd)

		// Find the actual git command
		var gitCommand string
		for i, arg := range args {
			if arg == "-C" && i+1 < len(args) {
				i++
				continue
			}
			if arg == "fetch" || arg == "checkout" {
				gitCommand = arg
				break
			}
		}

		switch gitCommand {
		case "fetch":
			return []byte("Fetched successfully"), nil // Simulate fetch success
		case "checkout":
			return mockOutput, mockError // Simulate checkout failure
		default:
			return nil, fmt.Errorf("unexpected git command structure: %v", args)
		}
	}

	// Act
	fetcherInstance := fetcher.GitFetcher{
		Runner: mockRunner,
		Logger: slog.New(slog.DiscardHandler),
	}
	err := fetcherInstance.Fetch(ctx, source, destDir)

	// Assert
	assert.Error(t, err, "Fetch should fail when git checkout fails")
	assert.Contains(t, err.Error(), "failed to checkout revision", "Error message should indicate checkout failure")
	assert.ErrorIs(t, err, mockError, "Error should wrap the runner error")
	assert.Contains(t, err.Error(), string(mockOutput), "Error message should contain output from command")
	require.Len(t, executedCommands, 2, "Expected two git commands")
	assert.Equal(t, []string{"git", "-C", destDir, "fetch", "origin"}, executedCommands[0])
	assert.Equal(t, []string{"git", "-C", destDir, "checkout", revision}, executedCommands[1])
}

// Add more tests below

func TestGitFetcher_Fetch_PullLatest(t *testing.T) {
	ctx := t.Context()
	destDir := t.TempDir() // Simulate existing directory
	source := domain.InputSource{
		Type: "git",
		Details: &domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			// Revision is empty, so pull latest
		},
	}

	// Arrange Mocks
	var executedCommands [][]string
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		require.Equal(t, "git", name)
		cmd := append([]string{name}, args...)
		executedCommands = append(executedCommands, cmd)

		// Find the actual git command (should be pull)
		var gitCommand string
		for i, arg := range args {
			if arg == "-C" && i+1 < len(args) {
				i++
				continue
			}
			if arg == "pull" {
				gitCommand = arg
				break
			}
		}

		if gitCommand == "pull" {
			return []byte("Already up to date."), nil // Simulate successful pull
		}
		return nil, fmt.Errorf("unexpected git command structure: %v", args)
	}

	// Act
	fetcherInstance := fetcher.GitFetcher{
		Runner: mockRunner,
		Logger: slog.New(slog.DiscardHandler),
	}
	err := fetcherInstance.Fetch(ctx, source, destDir)

	// Assert
	assert.NoError(t, err, "Fetch should succeed when pulling latest")
	require.Len(t, executedCommands, 1, "Expected one git command")
	assert.Equal(
		t,
		[]string{"git", "-C", destDir, "pull", "origin"},
		executedCommands[0],
		"Command should be git pull with -C",
	)
}

func TestGitFetcher_Fetch_PullLatest_Failure(t *testing.T) {
	ctx := t.Context()
	destDir := t.TempDir()
	source := domain.InputSource{
		Type: "git",
		Details: &domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
		},
	}

	// Arrange Mocks
	var executedCommands [][]string
	mockError := errors.New("mock git pull error")
	mockOutput := []byte("fatal: Could not read from remote repository.")
	mockRunner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		require.Equal(t, "git", name)
		cmd := append([]string{name}, args...)
		executedCommands = append(executedCommands, cmd)

		// Find the actual git command
		var gitCommand string
		for i, arg := range args {
			if arg == "-C" && i+1 < len(args) {
				i++
				continue
			}
			if arg == "pull" {
				gitCommand = arg
				break
			}
		}

		if gitCommand == "pull" {
			return mockOutput, mockError // Simulate pull failure
		}
		return nil, fmt.Errorf("unexpected git command structure: %v", args)
	}

	// Act
	fetcherInstance := fetcher.GitFetcher{
		Runner: mockRunner,
		Logger: slog.New(slog.DiscardHandler),
	}
	err := fetcherInstance.Fetch(ctx, source, destDir)

	// Assert
	assert.Error(t, err, "Fetch should fail when git pull fails")
	assert.Contains(t, err.Error(), "failed to pull latest changes", "Error message should indicate pull failure")
	assert.ErrorIs(t, err, mockError, "Error should wrap the runner error")
	assert.Contains(t, err.Error(), string(mockOutput), "Error message should contain output from command")
	require.Len(t, executedCommands, 1, "Expected one git command")
	assert.Equal(t, []string{"git", "-C", destDir, "pull", "origin"}, executedCommands[0])
}
