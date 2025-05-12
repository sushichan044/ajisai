package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/utils"
)

func TestResolveAbsPath(t *testing.T) {
	t.Run("absolute path is returned as is", func(t *testing.T) {
		// Consider Windows and Unix-like systems
		var absPath string
		if filepath.IsAbs("/absolute/path") {
			absPath = "/absolute/path"
		} else {
			// Windowsの場合はC:\path\format等
			absPath = "C:\\absolute\\path"
		}

		resolved, err := utils.ResolveAbsPath(absPath)
		require.NoError(t, err)
		assert.Equal(t, absPath, resolved)
	})

	t.Run("relative path is joined with working directory", func(t *testing.T) {
		wd, err := os.Getwd()
		require.NoError(t, err)

		relPath := "relative/path"
		expected := filepath.Join(wd, relPath)

		resolved, err := utils.ResolveAbsPath(relPath)
		require.NoError(t, err)
		assert.Equal(t, expected, resolved)
	})

	t.Run("empty path is resolved to working directory", func(t *testing.T) {
		wd, err := os.Getwd()
		require.NoError(t, err)

		resolved, err := utils.ResolveAbsPath("")
		require.NoError(t, err)
		assert.Equal(t, wd, resolved)
	})

	t.Run("tilde in path is expanded to home directory", func(t *testing.T) {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skip("Unable to determine home directory, skipping test")
		}

		tildeRelPath := "~/some/path"
		expected := filepath.Join(home, "some/path")

		resolved, err := utils.ResolveAbsPath(tildeRelPath)
		require.NoError(t, err)
		assert.Equal(t, expected, resolved)
	})
}
