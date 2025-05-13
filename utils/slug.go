package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

/*
GetSlugFromBaseDir returns the slug of the target path relative to the base directory.
Separator is always `/` (unix style).

	GetSlugFromBaseDir("/base", "/base/foo/bar/baz.tar.gz") // returns "foo/bar/baz.tar"
*/
func GetSlugFromBaseDir(baseAbsDir, targetPath string) (string, error) {
	cleanBaseAbsDir := filepath.Clean(baseAbsDir)
	cleanTargetPath := filepath.Clean(targetPath)

	// check if baseAbsDir points an absolute directory
	if !filepath.IsAbs(cleanBaseAbsDir) || filepath.Ext(cleanBaseAbsDir) != "" {
		return "", fmt.Errorf("baseAbsDir %s is not an absolute directory", baseAbsDir)
	}

	// check if targetPath points an absolute file
	if !filepath.IsAbs(cleanTargetPath) || filepath.Ext(cleanTargetPath) == "" {
		return "", fmt.Errorf("targetPath %s is not an absolute file", targetPath)
	}

	relPath, err := filepath.Rel(cleanBaseAbsDir, cleanTargetPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	if strings.HasPrefix(relPath, ".") {
		return "", fmt.Errorf("target path %s is not under base path %s", targetPath, baseAbsDir)
	}

	// ensure the path is normalized to unix style
	slashed := filepath.ToSlash(relPath)

	dir := filepath.Dir(slashed)
	base := filepath.Base(slashed)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	if dir == "." {
		return base, nil
	}

	return dir + "/" + base, nil
}
