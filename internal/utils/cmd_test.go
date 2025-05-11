package utils_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/aisync/internal/utils"
)

func TestDefaultCommandRunner_Run(t *testing.T) {
	runner := &utils.DefaultCommandRunner{}

	t.Run("successful command execution", func(t *testing.T) {
		// Use echo command as it's available on all platforms
		err := runner.Run("echo", "test")
		assert.NoError(t, err, "Run should not return an error for successful command")
	})

	t.Run("error on non-existent command", func(t *testing.T) {
		err := runner.Run("non_existent_command")
		require.Error(t, err, "Run should return an error for non-existent command")

		// Check the error type
		var cmdErr *utils.RunCommandError
		require.ErrorAs(t, err, &cmdErr)

		// Verify error contains command details
		assert.Contains(t, cmdErr.Error(), "non_existent_command", "Error message should contain the command name")
	})
}

func TestDefaultCommandRunner_RunInDir(t *testing.T) {
	runner := &utils.DefaultCommandRunner{}

	t.Run("execution in specified directory", func(t *testing.T) {
		tempDir := t.TempDir()

		marker := "testmarker"
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, marker), []byte(""), 0644))

		// Use marker file to check that the command was executed in the correct directory.
		err := runner.RunInDir(tempDir, "cat", marker)
		assert.NoError(t, err, "RunInDir should not return an error for successful command in existing dir")
	})

	t.Run("error on non-existent directory", func(t *testing.T) {
		nonExistentDir := filepath.Join(t.TempDir(), "non_existent_dir")
		// RunInDir itself doesn't check directory existence, exec.Command does when Run is called.
		err := runner.RunInDir(nonExistentDir, "echo", "test")

		require.Error(t, err, "RunInDir should return an error when the directory does not exist")

		// The specific error might depend on the OS and shell,
		// but it originates from the underlying exec call failing due to the directory.
		var runCmdErr *utils.RunCommandError
		require.ErrorAs(t, err, &runCmdErr)
		assert.ErrorContains(t, runCmdErr.CauseError, "no such file or directory") // Check underlying error
	})
}

func TestRunCommand(t *testing.T) {
	t.Run("successful command execution", func(t *testing.T) {
		cmd := exec.Command("echo", "test")
		err := utils.RunCommand(cmd)
		assert.NoError(t, err, "RunCommand should not return an error for successful command")
	})

	t.Run("error on failed command", func(t *testing.T) {
		// Fail!
		cmd := exec.Command("sh", "-c", "exit 1")
		err := utils.RunCommand(cmd)

		var cmdErr *utils.RunCommandError
		require.ErrorAs(t, err, &cmdErr)

		assert.Contains(t, cmdErr.Error(), "sh -c exit 1", "Error message should contain the command")
	})
}

func TestRunCommandError(t *testing.T) {
	t.Run("error formatting", func(t *testing.T) {
		cmd := exec.Command("test_cmd", "arg1", "arg2")
		causeErr := errors.New("test error")
		cmdErr := &utils.RunCommandError{
			Cmd:        cmd,
			CauseError: causeErr,
		}

		errMsg := cmdErr.Error()
		assert.Contains(t, errMsg, "test_cmd", "Error message should contain the command name")
		assert.Contains(t, errMsg, "test error", "Error message should contain the cause error")
	})

	t.Run("unwrap functionality", func(t *testing.T) {
		cmd := exec.Command("test_cmd")
		causeErr := errors.New("cause error")
		cmdErr := &utils.RunCommandError{
			Cmd:        cmd,
			CauseError: causeErr,
		}

		assert.ErrorIs(t, cmdErr, causeErr)
	})
}
