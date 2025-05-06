package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// CommandRunner defines the interface for running external commands.
type CommandRunner interface {
	Run(command string, args ...string) error
	RunInDir(dir string, command string, args ...string) error
}

// DefaultCommandRunner implements CommandRunner using os/exec.
type DefaultCommandRunner struct{}

var _ CommandRunner = (*DefaultCommandRunner)(nil)

// Run executes a command.
func (r *DefaultCommandRunner) Run(command string, args ...string) error {
	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return RunCommand(cmd)
}

// RunInDir executes a command in the specified directory.
func (r *DefaultCommandRunner) RunInDir(dir string, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return RunCommand(cmd)
}

func RunCommand(cmd *exec.Cmd) error {
	err := cmd.Run()
	if err != nil {
		return &RunCommandError{
			Cmd:        cmd,
			CauseError: err,
		}
	}
	return nil
}

type RunCommandError struct {
	Cmd        *exec.Cmd
	CauseError error
}

func (e *RunCommandError) Error() string {
	return fmt.Sprintf("%s: %s", e.Cmd.String(), e.CauseError)
}

func (e *RunCommandError) Unwrap() error {
	return e.CauseError
}
