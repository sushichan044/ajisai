package config

import (
	"fmt"
	"path/filepath"

	"github.com/sushichan044/ajisai/utils"
)

type Config struct {
	/*
		Tool-wide settings.

		Has no effect for package definition.
	*/
	Settings *Settings

	/*
		Definition to treat this workspace as a preset package.
	*/
	Package *Package

	/*
		Definition to use various presets in this workspace.
	*/
	Workspace *Workspace
}

func (c *Config) GetImportedPackageCacheRoot(packageName string) (string, error) {
	cacheDir, err := utils.ResolveAbsPath(c.Settings.CacheDir)
	if err != nil {
		return "", err
	}

	_, isConfigured := c.Workspace.Imports[packageName]
	if !isConfigured {
		return "", fmt.Errorf("package %s not found", packageName)
	}

	return filepath.Join(cacheDir, packageName), nil
}
