package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/aisync/utils"
)

func TestEmptyDir(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), "arm-test-empty-dir")
	require.NoError(t, os.MkdirAll(tempDir, 0750))
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0640))

	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0750))
	subFile := filepath.Join(subDir, "subfile.txt")
	require.NoError(t, os.WriteFile(subFile, []byte("sub content"), 0640))

	err := utils.EmptyDir(tempDir)
	require.NoError(t, err, "EmptyDir should not return an error")

	// removeAll removes tempDir and all its contents
	_, err = os.Stat(tempDir)
	assert.ErrorIs(t, err, os.ErrNotExist, "Directory should be removed")
}

func TestEnsureDir(t *testing.T) {
	t.Run("Create non-existent directory", func(t *testing.T) {
		tempDir := filepath.Join(t.TempDir(), "arm-test-ensure-dir")
		defer os.RemoveAll(tempDir)

		_, err := os.Stat(tempDir)
		require.ErrorIs(t, err, os.ErrNotExist, "Directory should not exist before test")

		err = utils.EnsureDir(tempDir)
		require.NoError(t, err, "EnsureDir should not return an error for non-existent directory")

		stat, err := os.Stat(tempDir)
		require.NoError(t, err, "Directory should exist after EnsureDir")
		assert.True(t, stat.IsDir(), "Created path should be a directory")
	})

	t.Run("Use existing directory", func(t *testing.T) {
		tempDir := filepath.Join(t.TempDir(), "arm-test-ensure-dir-existing")
		require.NoError(t, os.MkdirAll(tempDir, 0750))
		defer os.RemoveAll(tempDir)

		err := utils.EnsureDir(tempDir)
		assert.NoError(t, err, "EnsureDir should not return an error for existing directory")
	})

	t.Run("Path is a file", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "arm-test-ensure-dir-file")
		require.NoError(t, os.WriteFile(tempFile, []byte("test content"), 0640))
		defer os.Remove(tempFile)

		err := utils.EnsureDir(tempFile)
		require.Error(t, err, "EnsureDir should return an error when path is a file")
		assert.ErrorContains(t, err, "is not a directory", "Error message should mention path is not a directory")
	})
}

func TestIsDirExists(t *testing.T) {
	t.Run("Directory exists", func(t *testing.T) {
		dir := t.TempDir()
		defer os.RemoveAll(dir)

		exists, err := utils.IsDirExists(dir)
		require.NoError(t, err, "IsDirExists should not return an error")
		assert.True(t, exists, "Directory should exist")
	})

	t.Run("Directory does not exist", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "non-existent-dir")
		exists, err := utils.IsDirExists(dir)
		require.NoError(t, err, "IsDirExists should not return an error")
		assert.False(t, exists, "Directory should not exist")
	})

	t.Run("Path is a file", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "arm-test-is-dir-exists-file")
		require.NoError(t, os.WriteFile(file, []byte("test content"), 0640))
		defer os.Remove(file)

		exists, err := utils.IsDirExists(file)
		require.ErrorContains(t, err, "is not a directory")
		assert.False(t, exists, "File should not be considered a directory")
	})
}
