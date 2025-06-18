package loader

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/config"
	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type agentPresetLoader struct {
	cfg *config.Config
}

func NewAgentPresetPackageLoader(config *config.Config) domain.AgentPresetPackageLoader {
	return &agentPresetLoader{cfg: config}
}

func (l *agentPresetLoader) LoadAgentPresetPackage(packageName string) (*domain.AgentPresetPackage, error) {
	importedPkgCfg, isImported := l.cfg.Workspace.Imports[packageName]
	if !isImported {
		return nil, fmt.Errorf("could not load package %s: not imported", packageName)
	}

	pkgManifest, manifestErr := l.ResolvePackageManifest(packageName)
	if manifestErr != nil {
		return nil, fmt.Errorf("failed to load package manifest for %s: %w", packageName, manifestErr)
	}

	eg := errgroup.Group{}
	importedPresets := make([]*domain.AgentPreset, 0, len(importedPkgCfg.Include))

	for _, includedPresetName := range importedPkgCfg.Include {
		if _, isExported := pkgManifest.Exports[includedPresetName]; !isExported {
			// TODO: log warning
			continue
		}

		eg.Go(func() error {
			preset, buildErr := l.buildPreset(pkgManifest, includedPresetName)

			if buildErr != nil {
				return fmt.Errorf("build preset %s: %w", includedPresetName, buildErr)
			}
			importedPresets = append(importedPresets, preset)
			return nil
		})
	}

	if groupErr := eg.Wait(); groupErr != nil {
		return nil, fmt.Errorf("build agent preset package %s: %w", packageName, groupErr)
	}

	return &domain.AgentPresetPackage{
		PackageName: packageName,
		Presets:     importedPresets,
	}, nil
}

func (l *agentPresetLoader) ResolvePackageManifest(packageName string) (*config.Package, error) {
	cacheDir, err := l.cfg.GetImportedPackageCacheRoot(packageName)
	if err != nil {
		return nil, fmt.Errorf("resolve package manifest for %s: %w", packageName, err)
	}

	manager, err := config.NewDefaultManagerInDir(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("resolve package manifest for %s: %w", packageName, err)
	}

	rawManifest, err := manager.Load()
	if err != nil {
		var manifestNotFound *config.NoFileToReadError
		if errors.As(err, &manifestNotFound) {
			// manifest file not found, fallback to special `default` preset
			return &config.Package{
				Name: packageName,
				Exports: map[string]config.ExportedPresetDefinition{
					config.DefaultPresetName: {
						Prompts: []string{"prompts/**/*.md"},
						Rules:   []string{"rules/**/*.md"},
					},
				},
			}, nil
		}

		return nil, fmt.Errorf("resolve package manifest for %s: %w", packageName, err)
	}

	return &config.Package{
		Name:    packageName,
		Exports: rawManifest.Package.Exports,
	}, nil
}

// buildPreset scans the source directory for rules and prompts and returns a Preset.
func (l *agentPresetLoader) buildPreset(pkgManifest *config.Package, presetName string) (*domain.AgentPreset, error) {
	rootDir, err := l.cfg.GetImportedPackageCacheRoot(pkgManifest.Name)
	if err != nil {
		return nil, fmt.Errorf("build preset %s: %w", presetName, err)
	}
	exports, isExported := pkgManifest.Exports[presetName]
	if !isExported {
		return nil, fmt.Errorf("preset %s is not exported", presetName)
	}

	var (
		prompts []*domain.PromptItem
		rules   []*domain.RuleItem
	)

	eg := errgroup.Group{}

	for _, promptGlob := range exports.Prompts {
		eg.Go(func() error {
			loadedPrompts, loadErr := l.loadPromptItems(rootDir, pkgManifest.Name, presetName, promptGlob)
			if loadErr != nil {
				return fmt.Errorf("glob failed for prompt %s: %w", promptGlob, loadErr)
			}
			prompts = append(prompts, loadedPrompts...)
			return nil
		})
	}

	for _, ruleGlob := range exports.Rules {
		eg.Go(func() error {
			loadedRules, loadErr := l.loadRuleItems(rootDir, pkgManifest.Name, presetName, ruleGlob)
			if loadErr != nil {
				return fmt.Errorf("glob failed for rule %s: %w", ruleGlob, loadErr)
			}
			rules = append(rules, loadedRules...)
			return nil
		})
	}

	if groupErr := eg.Wait(); groupErr != nil {
		return nil, groupErr
	}

	return &domain.AgentPreset{
		Name:    presetName,
		Rules:   rules,
		Prompts: prompts,
	}, nil
}

func (l *agentPresetLoader) loadPromptItems(
	rootDir, packageName, presetName, promptGlob string,
) ([]*domain.PromptItem, error) {
	var loadedPrompts []*domain.PromptItem
	slashed := filepath.ToSlash(promptGlob)
	base, glob := doublestar.SplitPattern(slashed)
	fsys := os.DirFS(filepath.Join(rootDir, base))

	err := doublestar.GlobWalk(fsys, glob, func(path string, d fs.DirEntry) error {
		if d.IsDir() || !strings.HasSuffix(path, domain.PromptInternalExtension) {
			return nil
		}

		// Construct the full path relative to the actual file system for ReadFile
		// and GetSlugFromBaseDir, as `path` is relative to `fsys`'s root.
		fullPath := filepath.Join(rootDir, base, path)

		uriPath, pathErr := domain.GetPathFromBaseDir(filepath.Join(rootDir, base), fullPath)
		if pathErr != nil {
			return fmt.Errorf("failed to get path for prompt %s: %w", fullPath, pathErr)
		}

		body, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			return fmt.Errorf("failed to read prompt file %s: %w", fullPath, readErr)
		}

		result, parseErr := utils.ParseMarkdownWithMetadata[domain.PromptMetadata](body)
		if parseErr != nil {
			return fmt.Errorf("failed to parse prompt markdown %s: %w", fullPath, parseErr)
		}

		uri := domain.URI{
			Scheme:  domain.Scheme,
			Package: packageName,
			Preset:  presetName,
			Type:    domain.PromptsPresetType,
			Path:    uriPath,
		}
		promptItem := domain.NewPromptItem(uri, result.Content, result.FrontMatter)
		loadedPrompts = append(loadedPrompts, promptItem)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return loadedPrompts, nil
}

func (l *agentPresetLoader) loadRuleItems(
	rootDir, packageName, presetName, ruleGlob string,
) ([]*domain.RuleItem, error) {
	var loadedRules []*domain.RuleItem
	slashed := filepath.ToSlash(ruleGlob)
	base, glob := doublestar.SplitPattern(slashed)
	fsys := os.DirFS(filepath.Join(rootDir, base))

	err := doublestar.GlobWalk(fsys, glob, func(path string, d fs.DirEntry) error {
		if d.IsDir() || !strings.HasSuffix(path, domain.RuleInternalExtension) {
			return nil
		}

		// Construct the full path relative to the actual file system for ReadFile
		// and GetSlugFromBaseDir, as `path` is relative to `fsys`'s root.
		fullPath := filepath.Join(rootDir, base, path)

		uriPath, pathErr := domain.GetPathFromBaseDir(filepath.Join(rootDir, base), fullPath)
		if pathErr != nil {
			return fmt.Errorf("failed to get path for rule %s: %w", fullPath, pathErr)
		}

		body, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			return fmt.Errorf("failed to read rule file %s: %w", fullPath, readErr)
		}

		result, parseErr := utils.ParseMarkdownWithMetadata[domain.RuleMetadata](body)
		if parseErr != nil {
			return fmt.Errorf("failed to parse rule markdown %s: %w", fullPath, parseErr)
		}

		uri := domain.URI{
			Scheme:  domain.Scheme,
			Package: packageName,
			Preset:  presetName,
			Type:    domain.RulesPresetType,
			Path:    uriPath,
		}
		ruleItem := domain.NewRuleItem(uri, result.Content, result.FrontMatter)
		loadedRules = append(loadedRules, ruleItem)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return loadedRules, nil
}
