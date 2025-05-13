package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/internal/fetcher"
	"github.com/sushichan044/ajisai/internal/parser"
	"github.com/sushichan044/ajisai/internal/repository"
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
// Returns the preset names of the fetched presets.
func (engine *Engine) Fetch() ([]string, error) {
	eg := errgroup.Group{}

	presetNames := make([]string, 0, len(engine.cfg.Inputs))

	for identifier, input := range engine.cfg.Inputs {
		eg.Go(func() error {
			presetNames = append(presetNames, identifier)

			fetcher, err := getFetcher(input.Type)
			if err != nil {
				return fmt.Errorf("unknown input type: %s", input.Type)
			}

			return fetcher.Fetch(input, filepath.Join(engine.cfg.Settings.CacheDir, identifier))
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return presetNames, nil
}

// Parse parses the presets from the preset names and returns them.
func (engine *Engine) Parse(presetNames []string) ([]domain.AgentPreset, error) {
	eg := errgroup.Group{}

	presets := make([]domain.AgentPreset, 0, len(presetNames))

	for _, presetName := range presetNames {
		eg.Go(func() error {
			parsedPkg, parseErr := parser.ParsePreset(engine.cfg, presetName)
			if parseErr != nil {
				return fmt.Errorf("failed to parse preset: %w", parseErr)
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
			return repository.Clean(engine.cfg.Settings.Namespace)
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
func (engine *Engine) Export(presets []domain.AgentPreset) error {
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
		for _, preset := range presets {
			eg.Go(func() error {
				return currentRepo.WritePreset(engine.cfg.Settings.Namespace, preset)
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func getFetcher(inputType domain.PresetSourceType) (domain.ContentFetcher, error) {
	switch inputType {
	case domain.PresetSourceTypeLocal:
		return fetcher.NewLocalFetcher(), nil
	case domain.PresetSourceTypeGit:
		return fetcher.NewGitFetcher(), nil
	default:
		return nil, fmt.Errorf("unknown input type: %s", inputType)
	}
}

func getRepository(target domain.SupportedAgentType) (domain.PresetRepository, error) {
	switch target {
	case domain.SupportedAgentTypeCursor:
		return repository.NewPresetRepository(repository.NewCursorAdapter())
	case domain.SupportedAgentTypeGitHubCopilot:
		return repository.NewPresetRepository(repository.NewGitHubCopilotAdapter())
	case domain.SupportedAgentTypeWindsurf:
		return repository.NewPresetRepository(repository.NewWindsurfAdapter())
	default:
		return nil, fmt.Errorf("unknown output type: %s", target)
	}
}
