package engine

import (
	"fmt"
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

func NewEngine(cfg *domain.Config) *Engine {
	return &Engine{cfg: cfg}
}

func (engine *Engine) Fetch() ([]string, error) {
	eg := errgroup.Group{}

	packageNames := make([]string, 0, len(engine.cfg.Inputs))

	for identifier, input := range engine.cfg.Inputs {
		eg.Go(func() error {
			packageNames = append(packageNames, identifier)

			fetcher := getFetcher(input.Type)
			if fetcher == nil {
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

func (engine *Engine) Write(presets []domain.PresetPackage) error {
	eg := errgroup.Group{}

	enabledOutputs := make([]domain.OutputTarget, 0, len(engine.cfg.Outputs))
	for _, output := range engine.cfg.Outputs {
		if output.Enabled {
			enabledOutputs = append(enabledOutputs, output)
		}
	}

	for _, output := range enabledOutputs {
		repository := getRepository(output.Target)

		for _, pkg := range presets {
			eg.Go(func() error {
				return repository.WritePackage(engine.cfg.Global.Namespace, pkg)
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func getFetcher(inputType string) domain.ContentFetcher {
	switch inputType {
	case "local":
		return fetcher.LocalFetcher()
	case "git":
		return fetcher.GitFetcher()
	default:
		return nil
	}
}

func getRepository(target string) domain.PresetRepository {
	switch target {
	case "cursor":
		return repository.NewCursorRepository()
	default:
		return nil
	}
}
