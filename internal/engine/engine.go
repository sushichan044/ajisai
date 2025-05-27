package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/config"
	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/internal/fetcher"
	"github.com/sushichan044/ajisai/internal/integration"
	"github.com/sushichan044/ajisai/internal/loader"
)

type Engine struct {
	cfg *config.Config

	activeIntegrations []domain.AgentIntegration
}

func NewEngine(cfg *config.Config) (*Engine, error) {
	if cfg == nil {
		return nil, errors.New("internal error: config is nil")
	}

	activeIntegrations, integErr := getEnabledIntegrations(cfg)
	if integErr != nil {
		return nil, fmt.Errorf("failed to get enabled integrations: %w", integErr)
	}

	return &Engine{cfg: cfg, activeIntegrations: activeIntegrations}, nil
}

func (engine *Engine) ApplyPackage(packageName string) error {
	fetchErr := engine.fetchPackage(packageName)
	if fetchErr != nil {
		return fmt.Errorf("failed to fetch package %s: %w", packageName, fetchErr)
	}

	pkg, loadErr := engine.LoadPackage(packageName)

	if loadErr != nil {
		return fmt.Errorf("failed to load package %s: %w", packageName, loadErr)
	}

	exportErr := engine.exportPackage(pkg)
	if exportErr != nil {
		return fmt.Errorf("failed to export package %s: %w", packageName, exportErr)
	}

	return nil
}

func (engine *Engine) CleanOutputs() error {
	eg := errgroup.Group{}

	for _, integration := range engine.activeIntegrations {
		eg.Go(func() error {
			return integration.Clean(engine.cfg.Settings.Namespace)
		})
	}

	return eg.Wait()
}

func (engine *Engine) CleanCache(force bool) error {
	cacheDir := engine.cfg.Settings.CacheDir

	if _, err := os.Stat(cacheDir); errors.Is(err, os.ErrNotExist) {
		// Nothing to clean.
		return nil
	}

	if force {
		// Remove all cache directories.
		if err := os.RemoveAll(cacheDir); err != nil {
			return fmt.Errorf("failed to remove cache directory %s: %w", cacheDir, err)
		}

		// Create the cache directory again.
		return os.MkdirAll(cacheDir, 0750)
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory %s: %w", cacheDir, err)
	}

	eg := errgroup.Group{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip files, only interested in preset directories
		}
		entryName := entry.Name()

		eg.Go(func() error {
			if _, isConfigured := engine.cfg.Workspace.Imports[entryName]; !isConfigured {
				// This entry is not configured in inputs, so we don't need it.
				pathToRemove := filepath.Join(cacheDir, entryName)
				if removeErr := os.RemoveAll(pathToRemove); removeErr != nil {
					return fmt.Errorf("failed to remove obsolete cache directory %s: %w", pathToRemove, removeErr)
				}
			}

			return nil
		})
	}

	return eg.Wait()
}

func (engine *Engine) fetchPackage(packageName string) error {
	pkgImport, imported := engine.cfg.Workspace.Imports[packageName]
	if !imported {
		return fmt.Errorf("package %s is not imported", packageName)
	}

	fetcher, fetcherBuildErr := getFetcher(pkgImport.Type)
	if fetcherBuildErr != nil {
		return fmt.Errorf("failed to get fetcher: %w", fetcherBuildErr)
	}

	cacheDestination, cacheErr := engine.cfg.GetImportedPackageCacheRoot(packageName)
	if cacheErr != nil {
		return fmt.Errorf("failed to get cache root for package %s: %w", packageName, cacheErr)
	}

	return fetcher.Fetch(pkgImport, cacheDestination)
}

func (engine *Engine) LoadPackage(packageName string) (*domain.AgentPresetPackage, error) {
	loader := loader.NewAgentPresetPackageLoader(engine.cfg)
	return loader.LoadAgentPresetPackage(packageName)
}

func (engine *Engine) exportPackage(pkg *domain.AgentPresetPackage) error {
	eg := errgroup.Group{}

	for _, integration := range engine.activeIntegrations {
		eg.Go(func() error {
			return integration.WritePackage(engine.cfg.Settings.Namespace, pkg)
		})
	}

	return eg.Wait()
}

func getFetcher(inputType config.ImportType) (domain.PackageFetcher, error) {
	switch inputType {
	case config.ImportTypeLocal:
		return fetcher.NewLocalFetcher(), nil
	case config.ImportTypeGit:
		return fetcher.NewGitFetcher(), nil
	}
	return nil, fmt.Errorf("unknown import type: %s", inputType)
}

func getEnabledIntegrations(cfg *config.Config) ([]domain.AgentIntegration, error) {
	maxIntegrations := 3
	integrations := make([]domain.AgentIntegration, 0, maxIntegrations)

	// Check if Integrations is nil to avoid nil pointer dereference
	if cfg.Workspace == nil || cfg.Workspace.Integrations == nil {
		return integrations, nil
	}

	if cfg.Workspace.Integrations.Cursor != nil && cfg.Workspace.Integrations.Cursor.Enabled {
		cursorRepo, cursorErr := getIntegration(config.AgentIntegrationTypeCursor)
		if cursorErr != nil {
			return nil, fmt.Errorf("failed to get cursor repository: %w", cursorErr)
		}
		integrations = append(integrations, cursorRepo)
	}

	if cfg.Workspace.Integrations.GitHubCopilot != nil && cfg.Workspace.Integrations.GitHubCopilot.Enabled {
		githubCopilotRepo, githubCopilotErr := getIntegration(config.AgentIntegrationTypeGitHubCopilot)
		if githubCopilotErr != nil {
			return nil, fmt.Errorf("failed to get github copilot repository: %w", githubCopilotErr)
		}
		integrations = append(integrations, githubCopilotRepo)
	}

	if cfg.Workspace.Integrations.Windsurf != nil && cfg.Workspace.Integrations.Windsurf.Enabled {
		windsurfRepo, windsurfErr := getIntegration(config.AgentIntegrationTypeWindsurf)
		if windsurfErr != nil {
			return nil, fmt.Errorf("failed to get windsurf repository: %w", windsurfErr)
		}
		integrations = append(integrations, windsurfRepo)
	}

	return integrations, nil
}

func getIntegration(target config.AgentIntegrationType) (domain.AgentIntegration, error) {
	switch target {
	case config.AgentIntegrationTypeCursor:
		return integration.New(integration.NewCursorAdapter())
	case config.AgentIntegrationTypeGitHubCopilot:
		return integration.New(integration.NewGitHubCopilotAdapter())
	case config.AgentIntegrationTypeWindsurf:
		return integration.New(integration.NewWindsurfAdapter())
	}
	return nil, fmt.Errorf("unknown agent integration type: %s", target)
}
