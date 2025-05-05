package fetcher_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/fetcher"
)

func TestLocalFetcher_Fetch_SourceNotExist(t *testing.T) {
	fetcher := fetcher.NewLocalFetcher()
	tempDir := t.TempDir()
	nonExistentSourcePath := filepath.Join(tempDir, "non-existent-src")
	destPath := filepath.Join(tempDir, "dest")

	source := domain.InputSource{
		Type: "local",
		Details: domain.LocalInputSourceDetails{
			Path: nonExistentSourcePath,
		},
	}

	// --- Act
	err := fetcher.Fetch(context.Background(), source, destPath)

	// --- Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "source directory")
	assert.ErrorContains(t, err, "does not exist")
	assert.ErrorIs(t, err, os.ErrNotExist) // Check underlying error
}

func TestLocalFetcher_Fetch_SourceIsFile(t *testing.T) {
	fetcher := fetcher.NewLocalFetcher()
	tempDir := t.TempDir()
	sourceFilePath := filepath.Join(tempDir, "source.txt")
	destPath := filepath.Join(tempDir, "dest")

	// Create a source file
	require.NoError(t, os.WriteFile(sourceFilePath, []byte("hello"), 0644))

	source := domain.InputSource{
		Type: "local",
		Details: domain.LocalInputSourceDetails{
			Path: sourceFilePath,
		},
	}

	// --- Act
	err := fetcher.Fetch(context.Background(), source, destPath)

	// --- Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "is not a directory")
}

func TestLocalFetcher_Fetch_WrongDetailsType(t *testing.T) {
	fetcher := fetcher.NewLocalFetcher()
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "dest")

	source := domain.InputSource{
		Type: "git", // Wrong type for LocalFetcher
		Details: domain.GitInputSourceDetails{
			Repository: "some-repo",
		},
	}

	// --- Act
	err := fetcher.Fetch(context.Background(), source, destPath)

	// --- Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "unexpected source details type")
}

func TestLocalFetcher_Fetch_DestinationHandling(t *testing.T) {
	fetcher := fetcher.NewLocalFetcher()
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "src")
	destPath := filepath.Join(tempDir, "dest")

	// Create source directory
	require.NoError(t, os.MkdirAll(sourcePath, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourcePath, "file1.txt"), []byte("source"), 0644))

	source := domain.InputSource{
		Type:    "local",
		Details: domain.LocalInputSourceDetails{Path: sourcePath},
	}

	tests := []struct {
		name        string
		setupDest   func(t *testing.T)
		expectError bool
		assertAfter func(t *testing.T)
	}{
		{
			name: "destination does not exist",
			setupDest: func(t *testing.T) {
				// No setup needed, destPath initially doesn't exist
			},
			expectError: false, // Should succeed now
			assertAfter: func(t *testing.T) {
				// Check if dest dir was created
				info, err := os.Stat(destPath)
				require.NoError(t, err, "Destination directory should have been created")
				assert.True(t, info.IsDir(), "Destination should be a directory")
			},
		},
		{
			name: "destination is a file",
			setupDest: func(t *testing.T) {
				require.NoError(t, os.WriteFile(destPath, []byte("i am a file"), 0644))
			},
			expectError: true, // Still expect error
			assertAfter: func(t *testing.T) {
				// Check the file is still there
				info, err := os.Stat(destPath)
				require.NoError(t, err)
				assert.False(t, info.IsDir(), "Destination should remain a file")
			},
		},
		{
			name: "destination exists and is cleared",
			setupDest: func(t *testing.T) {
				require.NoError(t, os.MkdirAll(destPath, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(destPath, "existing.txt"), []byte("old"), 0644))
			},
			expectError: false, // Should succeed now
			assertAfter: func(t *testing.T) {
				// Check the directory exists and *will contain the copied file* (file1.txt)
				// We don't check for emptiness anymore, but rather the presence of the copied file.
				copiedFilePath := filepath.Join(destPath, "file1.txt")
				require.FileExists(t, copiedFilePath, "Copied file should exist in cleared directory")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset destination state for each test
			_ = os.RemoveAll(destPath)
			tc.setupDest(t)

			// --- Act
			err := fetcher.Fetch(context.Background(), source, destPath)

			// --- Assert
			if tc.expectError {
				require.Error(t, err)
				if tc.name == "destination is a file" {
					assert.ErrorContains(t, err, "not a directory")
				} else {
					// For other cases, we no longer expect a specific error,
					// just that *an* error occurred if expectError is true.
					// Remove the specific check for "copying TODO"
				}
			} else {
				require.NoError(t, err)
			}
			if tc.assertAfter != nil {
				tc.assertAfter(t)
			}
		})
	}
}

func TestLocalFetcher_Fetch_CopySuccess(t *testing.T) {
	// --- Setup Source ---
	sourceDir, err := os.MkdirTemp("", "fetcher_test_source_*")
	require.NoError(t, err)
	defer os.RemoveAll(sourceDir) // Clean up source

	// Create some source files and directories
	subDir := filepath.Join(sourceDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644)
	require.NoError(t, err)

	// --- Setup Destination ---
	destDir, err := os.MkdirTemp("", "fetcher_test_dest_*")
	require.NoError(t, err)
	defer os.RemoveAll(destDir) // Clean up destination

	// --- Execute Fetch ---
	fetcher := fetcher.NewLocalFetcher()
	inputSource := domain.InputSource{
		Type: "local",
		Details: domain.LocalInputSourceDetails{
			Path: sourceDir,
		},
	}
	err = fetcher.Fetch(context.Background(), inputSource, destDir)

	// --- Assertions ---
	require.NoError(t, err)

	// Verify file1 exists in destination
	destFile1Path := filepath.Join(destDir, "file1.txt")
	require.FileExists(t, destFile1Path)
	content1, err := os.ReadFile(destFile1Path)
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content1))

	// Verify subdir and file2 exist in destination
	destSubDirPath := filepath.Join(destDir, "subdir")
	require.DirExists(t, destSubDirPath)
	destFile2Path := filepath.Join(destSubDirPath, "file2.txt")
	require.FileExists(t, destFile2Path)
	content2, err := os.ReadFile(destFile2Path)
	require.NoError(t, err)
	assert.Equal(t, "content2", string(content2))
}
