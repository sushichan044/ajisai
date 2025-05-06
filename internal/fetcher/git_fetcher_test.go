package fetcher_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/fetcher"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

type MockCommandRunner struct {
	mock.Mock
}

func (m *MockCommandRunner) Run(command string, args ...string) error {
	callArgs := m.Called(command, args)
	return callArgs.Error(0)
}

func (m *MockCommandRunner) RunInDir(dir string, command string, args ...string) error {
	callArgs := m.Called(dir, command, args)
	return callArgs.Error(0)
}

func stubIsDirExists(t *testing.T, exists bool, err error) func() {
	original := utils.IsDirExists
	utils.IsDirExists = func(_ string) (bool, error) {
		return exists, err
	}
	t.Cleanup(func() {
		utils.IsDirExists = original
	})
	return func() { utils.IsDirExists = original }
}

func TestGitFetcher_Fetch_InitialClone(t *testing.T) {
	destDir := "/tmp/dest"
	absoluteDestDir, _ := filepath.Abs(destDir) // Assume Abs works
	repoURL := "https://github.com/example/repo.git"
	source := domain.InputSource{
		Type:    "git",
		Details: domain.GitInputSourceDetails{Repository: repoURL},
	}

	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)

	stubIsDirExists(t, false, nil) // Simulate directory does not exist

	expectedArgs := []string{"clone", repoURL, absoluteDestDir}
	mockRunner.On("Run", "git", expectedArgs).Return(nil)

	err := fetcherInstance.Fetch(source, destDir)

	require.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

func TestGitFetcher_Fetch_InitialClone_Failure(t *testing.T) {
	destDir := "/tmp/dest"
	absoluteDestDir, _ := filepath.Abs(destDir)
	repoURL := "invalid-url"
	source := domain.InputSource{
		Type:    "git",
		Details: domain.GitInputSourceDetails{Repository: repoURL},
	}

	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)

	stubIsDirExists(t, false, nil)

	cloneErr := errors.New("git clone failed")
	expectedArgs := []string{"clone", repoURL, absoluteDestDir}
	mockRunner.On("Run", "git", expectedArgs).Return(cloneErr)

	// Act
	err := fetcherInstance.Fetch(source, destDir)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, cloneErr)
	mockRunner.AssertExpectations(t)
}

func TestGitFetcher_Fetch_CheckoutRevision(t *testing.T) {
	destDir := "/tmp/existing-repo"
	absoluteDestDir, _ := filepath.Abs(destDir)
	revision := "v1.0.0"
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)

	stubIsDirExists(t, true, nil) // Simulate directory exists

	expectedFetchArgs := []string{"fetch", "origin"}
	mockRunner.On("RunInDir", absoluteDestDir, "git", expectedFetchArgs).Return(nil)

	expectedCheckoutArgs := []string{"checkout", revision}
	mockRunner.On("RunInDir", absoluteDestDir, "git", expectedCheckoutArgs).Return(nil)

	err := fetcherInstance.Fetch(source, destDir)

	require.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

func TestGitFetcher_Fetch_CheckoutRevision_FetchFailure(t *testing.T) {
	destDir := "/tmp/existing-repo"
	absoluteDestDir, _ := filepath.Abs(destDir)
	revision := "v1.0.0"
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)

	stubIsDirExists(t, true, nil)

	fetchErr := errors.New("git fetch failed")
	expectedFetchArgs := []string{"fetch", "origin"}
	mockRunner.On("RunInDir", absoluteDestDir, "git", expectedFetchArgs).Return(fetchErr)

	// Act
	err := fetcherInstance.Fetch(source, destDir)

	// Assert
	require.ErrorContains(t, err, "failed to fetch updates")
	require.ErrorIs(t, err, fetchErr)
	mockRunner.AssertExpectations(t)
	mockRunner.AssertNotCalled(
		t,
		"RunInDir",
		absoluteDestDir,
		"git",
		[]string{"checkout", revision},
	) // Ensure checkout wasn't called
}

func TestGitFetcher_Fetch_CheckoutRevision_CheckoutFailure(t *testing.T) {
	destDir := "/tmp/existing-repo"
	absoluteDestDir, _ := filepath.Abs(destDir)
	revision := "invalid-revision"
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)

	stubIsDirExists(t, true, nil)

	expectedFetchArgs := []string{"fetch", "origin"}
	mockRunner.On("RunInDir", absoluteDestDir, "git", expectedFetchArgs).Return(nil)

	checkoutErr := errors.New("git checkout failed")
	expectedCheckoutArgs := []string{"checkout", revision}
	mockRunner.On("RunInDir", absoluteDestDir, "git", expectedCheckoutArgs).Return(checkoutErr)

	// Act
	err := fetcherInstance.Fetch(source, destDir)

	// Assert
	require.ErrorIs(t, err, checkoutErr)
	mockRunner.AssertExpectations(t)
}

func TestGitFetcher_Fetch_PullLatest(t *testing.T) {
	destDir := "/tmp/existing-repo-pull"
	absoluteDestDir, _ := filepath.Abs(destDir)
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			// Revision is empty, so pull latest
		},
	}

	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)

	stubIsDirExists(t, true, nil) // Simulate directory exists

	expectedPullArgs := []string{"pull", "origin"}
	mockRunner.On("RunInDir", absoluteDestDir, "git", expectedPullArgs).Return(nil)

	err := fetcherInstance.Fetch(source, destDir)

	require.NoError(t, err)
	mockRunner.AssertExpectations(t)
}

func TestGitFetcher_Fetch_PullLatest_Failure(t *testing.T) {
	destDir := "/tmp/existing-repo-pull-fail"
	absoluteDestDir, _ := filepath.Abs(destDir)
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
		},
	}

	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)

	stubIsDirExists(t, true, nil)

	pullErr := errors.New("git pull failed")
	expectedPullArgs := []string{"pull", "origin"}
	mockRunner.On("RunInDir", absoluteDestDir, "git", expectedPullArgs).Return(pullErr)

	err := fetcherInstance.Fetch(source, destDir)

	require.ErrorIs(t, err, pullErr)
	mockRunner.AssertExpectations(t)
}

func TestGitFetcher_InvalidSourceType(t *testing.T) {
	mockRunner := new(MockCommandRunner) // Not actually used, but needed for constructor
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)
	destPath := "/tmp/dest"

	// Create a local input source (not git)
	source := domain.InputSource{
		Type: "local",
		Details: domain.LocalInputSourceDetails{
			Path: "/some/path",
		},
	}

	// Execute
	err := fetcherInstance.Fetch(source, destPath)

	// Assert
	require.Error(t, err)
	var invalidTypeErr *fetcher.InvalidSourceTypeError
	require.ErrorAs(t, err, &invalidTypeErr)
	assert.Equal(t, "git", invalidTypeErr.ExpectedType())
	assert.Equal(t, "local", invalidTypeErr.ActualType())
	assert.Contains(t, err.Error(), "expected source type: git, got: local")
}

func TestGitFetcher_EmptyRepository(t *testing.T) {
	mockRunner := new(MockCommandRunner)
	fetcherInstance := fetcher.GitFetcherWithRunner(mockRunner)
	destPath := "/tmp/dest"

	// Create a git input source with empty repository
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "", // Empty repo
		},
	}

	// Execute
	err := fetcherInstance.Fetch(source, destPath)

	// Verify
	require.Error(t, err)
	assert.EqualError(t, err, "git repository URL cannot be empty")
}
