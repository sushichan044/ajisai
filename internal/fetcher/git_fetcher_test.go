package fetcher_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomock "go.uber.org/mock/gomock"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/internal/fetcher"
	utils "github.com/sushichan044/ajisai/utils/mocks"
)

func TestGitFetcher_Fetch_InitialClone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDir := t.TempDir()
	destDir := filepath.Join(testDir, "dest")
	absoluteDestDir, err := filepath.Abs(destDir)
	require.NoError(t, err)

	repoURL := "https://github.com/example/repo.git"
	source := domain.InputSource{
		Type:    "git",
		Details: domain.GitInputSourceDetails{Repository: repoURL},
	}

	mockRunner := utils.NewMockCommandRunner(ctrl)
	expectedArgs := []any{"git", []string{"clone", repoURL, absoluteDestDir}}
	mockRunner.EXPECT().Run("git", gomock.Eq(expectedArgs[1])).Return(nil)

	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)

	err = fetcherInstance.Fetch(source, destDir)

	require.NoError(t, err)
}

func TestGitFetcher_Fetch_InitialClone_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)

	testDir := t.TempDir()
	destDir := filepath.Join(testDir, "dest")
	absoluteDestDir, err := filepath.Abs(destDir)
	require.NoError(t, err)
	repoURL := "invalid-url"
	source := domain.InputSource{
		Type:    "git",
		Details: domain.GitInputSourceDetails{Repository: repoURL},
	}

	cloneErr := errors.New("git clone failed")
	expectedArgs := []any{"git", []string{"clone", repoURL, absoluteDestDir}}
	mockRunner.EXPECT().Run("git", gomock.Eq(expectedArgs[1])).Return(cloneErr)

	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)
	err = fetcherInstance.Fetch(source, destDir)

	require.Error(t, err)
	require.ErrorIs(t, err, cloneErr)
}

func TestGitFetcher_Fetch_CheckoutRevision(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)

	testDir := t.TempDir()
	destDir := filepath.Join(testDir, "existing-repo")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	absoluteDestDir, err := filepath.Abs(destDir)
	require.NoError(t, err)

	revision := "v1.0.0"
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)

	expectedFetchArgs := []any{"git", []string{"fetch", "origin"}}
	mockRunner.EXPECT().RunInDir(absoluteDestDir, "git", gomock.Eq(expectedFetchArgs[1])).Return(nil)

	expectedCheckoutArgs := []any{"git", []string{"checkout", revision}}
	mockRunner.EXPECT().RunInDir(absoluteDestDir, "git", gomock.Eq(expectedCheckoutArgs[1])).Return(nil)

	err = fetcherInstance.Fetch(source, destDir)

	require.NoError(t, err)
}

func TestGitFetcher_Fetch_CheckoutRevision_FetchFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)

	testDir := t.TempDir()
	destDir := filepath.Join(testDir, "existing-repo")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	absoluteDestDir, err := filepath.Abs(destDir)
	require.NoError(t, err)

	revision := "v1.0.0"
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)

	fetchErr := errors.New("git fetch failed")
	expectedFetchArgs := []any{"git", []string{"fetch", "origin"}}
	mockRunner.EXPECT().RunInDir(absoluteDestDir, "git", gomock.Eq(expectedFetchArgs[1])).Return(fetchErr)

	err = fetcherInstance.Fetch(source, destDir)

	require.ErrorContains(t, err, "failed to fetch updates")
	require.ErrorIs(t, err, fetchErr)
}

func TestGitFetcher_Fetch_CheckoutRevision_CheckoutFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)

	testDir := t.TempDir()
	destDir := filepath.Join(testDir, "existing-repo")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	absoluteDestDir, err := filepath.Abs(destDir)
	require.NoError(t, err)

	revision := "invalid-revision"
	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
			Revision:   revision,
		},
	}

	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)

	expectedFetchArgs := []any{"git", []string{"fetch", "origin"}}
	mockRunner.EXPECT().RunInDir(absoluteDestDir, "git", gomock.Eq(expectedFetchArgs[1])).Return(nil)

	checkoutErr := errors.New("git checkout failed")
	expectedCheckoutArgs := []any{"git", []string{"checkout", revision}}
	mockRunner.EXPECT().RunInDir(absoluteDestDir, "git", gomock.Eq(expectedCheckoutArgs[1])).Return(checkoutErr)

	err = fetcherInstance.Fetch(source, destDir)

	require.ErrorIs(t, err, checkoutErr)
}

func TestGitFetcher_Fetch_PullLatest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)

	testDir := t.TempDir()
	destDir := filepath.Join(testDir, "existing-repo-pull")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	absoluteDestDir, err := filepath.Abs(destDir)
	require.NoError(t, err)

	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
		},
	}

	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)

	expectedPullArgs := []any{"git", []string{"pull", "origin"}}
	mockRunner.EXPECT().RunInDir(absoluteDestDir, "git", gomock.Eq(expectedPullArgs[1])).Return(nil)

	err = fetcherInstance.Fetch(source, destDir)

	require.NoError(t, err)
}

func TestGitFetcher_Fetch_PullLatest_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)

	testDir := t.TempDir()
	destDir := filepath.Join(testDir, "existing-repo-pull-fail")
	require.NoError(t, os.MkdirAll(destDir, 0755))
	absoluteDestDir, err := filepath.Abs(destDir)
	require.NoError(t, err)

	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "https://irrelevant.for/this/test",
		},
	}

	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)

	pullErr := errors.New("git pull failed")
	expectedPullArgs := []any{"git", []string{"pull", "origin"}}
	mockRunner.EXPECT().RunInDir(absoluteDestDir, "git", gomock.Eq(expectedPullArgs[1])).Return(pullErr)

	err = fetcherInstance.Fetch(source, destDir)

	require.ErrorIs(t, err, pullErr)
}

func TestGitFetcher_InvalidSourceType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)
	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)
	destPath := "/tmp/dest"

	source := domain.InputSource{
		Type: "local",
		Details: domain.LocalInputSourceDetails{
			Path: "/some/path",
		},
	}

	err := fetcherInstance.Fetch(source, destPath)

	require.Error(t, err)
	var invalidTypeErr *fetcher.InvalidSourceTypeError
	require.ErrorAs(t, err, &invalidTypeErr)
	assert.Equal(t, domain.InputSourceTypeGit, invalidTypeErr.ExpectedType())
	assert.Equal(t, domain.InputSourceTypeLocal, invalidTypeErr.ActualType())
	assert.Contains(t, err.Error(), "expected source type: git, got: local")
}

func TestGitFetcher_EmptyRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := utils.NewMockCommandRunner(ctrl)
	fetcherInstance := fetcher.NewGitFetcherWithRunner(mockRunner)
	destPath := "/tmp/dest"

	source := domain.InputSource{
		Type: "git",
		Details: domain.GitInputSourceDetails{
			Repository: "",
		},
	}

	err := fetcherInstance.Fetch(source, destPath)

	require.Error(t, err)
	assert.EqualError(t, err, "git repository URL cannot be empty")
}
