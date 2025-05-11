package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/fetcher"
	"github.com/sushichan044/ai-rules-manager/internal/parser"
	"github.com/sushichan044/ai-rules-manager/internal/repository"
)

type Engine struct {
	cfg *domain.Config
}

func NewEngine(cfg *domain.Config) (*Engine, error) {
	if cfg == nil {
		return nil, errors.New("internal error: config is nil")
	}

	return &Engine{cfg: cfg}, nil
}

// Fetch fetches presets from inputs and persist them in the cache directory.
// Returns the package names of the fetched presets.
func (engine *Engine) Fetch() ([]string, error) {
	eg := errgroup.Group{}

	packageNames := make([]string, 0, len(engine.cfg.Inputs))

	for identifier, input := range engine.cfg.Inputs {
		eg.Go(func() error {
			packageNames = append(packageNames, identifier)

			fetcher, err := getFetcher(input.Type)
			if err != nil {
				return fmt.Errorf("unknown input type: %s", input.Type)
			}

			return fetcher.Fetch(input, filepath.Join(engine.cfg.Global.CacheDir, identifier))
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return packageNames, nil
}

// Parse parses the presets from the package names and returns the preset packages.
func (engine *Engine) Parse(packageNames []string) ([]domain.PresetPackage, error) {
	eg := errgroup.Group{}

	presets := make([]domain.PresetPackage, 0, len(packageNames))

	for _, pkgName := range packageNames {
		eg.Go(func() error {
			parsedPkg, parseErr := parser.ParsePresetPackage(engine.cfg, pkgName)
			if parseErr != nil {
				return fmt.Errorf("failed to parse preset package: %w", parseErr)
			}

			presets = append(presets, *parsedPkg)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return presets, nil
}

func (engine *Engine) CleanOutputs() error {
	eg := errgroup.Group{}

	for _, output := range engine.cfg.Outputs {
		repository, err := getRepository(output.Target)
		if err != nil {
			return fmt.Errorf("unknown output type: %s", output.Target)
		}

		eg.Go(func() error {
			return repository.Clean(engine.cfg.Global.Namespace)
		})
	}

	return eg.Wait()
}

func (engine *Engine) CleanCache(force bool) error {
	cacheDir := engine.cfg.Global.CacheDir

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
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
			continue // Skip files, only interested in package directories
		}
		entryName := entry.Name()

		eg.Go(func() error {
			if _, isConfigured := engine.cfg.Inputs[entryName]; !isConfigured {
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

// Export exports the presets for specific agents configured in the outputs.
func (engine *Engine) Export(presets []domain.PresetPackage) error {
	repos := make([]domain.PresetRepository, 0, len(engine.cfg.Outputs))
	for _, output := range engine.cfg.Outputs {
		if !output.Enabled {
			continue
		}

		repo, err := getRepository(output.Target)
		if err != nil {
			return err
		}
		repos = append(repos, repo)
	}

	eg := errgroup.Group{}

	for _, currentRepo := range repos {
		for _, pkg := range presets {
			eg.Go(func() error {
				return currentRepo.WritePackage(engine.cfg.Global.Namespace, pkg)
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func getFetcher(inputType string) (domain.ContentFetcher, error) {
	switch inputType {
	case "local":
		return fetcher.NewLocalFetcher(), nil
	case "git":
		return fetcher.NewGitFetcher(), nil
	default:
		return nil, fmt.Errorf("unknown input type: %s", inputType)
	}
}

func getRepository(target string) (domain.PresetRepository, error) {
	switch target {
	case "cursor":
		return repository.NewCursorRepository(), nil
	case "github-copilot":
		return repository.NewGitHubCopilotRepository(), nil
	default:
		return nil, fmt.Errorf("unknown output type: %s", target)
	}
}
