package utils

import (
	"fmt"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

func IsPathGitIgnored(path string) (bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return false, fmt.Errorf("failed to get current working directory: %w", err)
	}

	ignore, compileErr := gitignore.CompileIgnoreFile(filepath.Join(cwd, ".gitignore"))
	if compileErr != nil {
		return false, fmt.Errorf("failed to compile gitignore: %w", compileErr)
	}

	return ignore.MatchesPath(path), nil
}
