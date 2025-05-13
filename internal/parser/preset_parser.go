package parser

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

// ParsePresetPackage scans the source directory for rules and prompts and returns a PresetPackage.
func ParsePresetPackage(config *domain.Config, presetName string) (*domain.PresetPackage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	if _, ok := config.Inputs[presetName]; !ok {
		return nil, fmt.Errorf("preset %s not found in config", presetName)
	}

	presetCacheRoot, err := config.GetPresetRootInCache(presetName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve preset root directory: %w", err)
	}

	var (
		prompts []*domain.PromptItem
		rules   []*domain.RuleItem
	)

	eg := new(errgroup.Group)

	eg.Go(func() error {
		parsedPrompts, innerErr := parsePrompts(presetCacheRoot)
		if innerErr != nil {
			return fmt.Errorf("failed to parse prompts: %w", innerErr)
		}

		prompts = parsedPrompts
		return nil
	})

	eg.Go(func() error {
		parsedRules, innerErr := parseRules(presetCacheRoot)
		if innerErr != nil {
			return fmt.Errorf("failed to parse rules: %w", innerErr)
		}

		rules = parsedRules
		return nil
	})

	if groupErr := eg.Wait(); groupErr != nil {
		return nil, groupErr
	}

	return &domain.PresetPackage{
		Name:    presetName,
		Rules:   rules,
		Prompts: prompts,
	}, nil
}

func parsePrompts(rootDir string) ([]*domain.PromptItem, error) {
	promptRootDir := filepath.Join(rootDir, string(domain.PromptsPresetType))
	items := []*domain.PromptItem{}

	if exists, err := utils.IsDirExists(promptRootDir); err != nil {
		return nil, fmt.Errorf("failed to check if prompt directory exists: %w", err)
	} else if !exists {
		return items, nil
	}

	walkErr := filepath.WalkDir(promptRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), "."+domain.PromptInternalExtension) {
			return nil
		}

		slug, err := utils.GetSlugFromBaseDir(promptRootDir, path)
		if err != nil {
			return err
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		result, err := utils.ParseMarkdownWithMetadata[domain.PromptMetadata](body)
		if err != nil {
			return err
		}

		ruleItem := domain.NewPromptItem(slug, result.Content, result.FrontMatter)

		items = append(items, ruleItem)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return items, nil
}

func parseRules(rootDir string) ([]*domain.RuleItem, error) {
	ruleRootDir := filepath.Join(rootDir, "rules")
	items := []*domain.RuleItem{}

	if exists, err := utils.IsDirExists(ruleRootDir); err != nil {
		return nil, fmt.Errorf("failed to check if rule directory exists: %w", err)
	} else if !exists {
		return items, nil
	}

	walkErr := filepath.WalkDir(ruleRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), "."+domain.RuleInternalExtension) {
			return nil
		}

		slug, err := utils.GetSlugFromBaseDir(ruleRootDir, path)
		if err != nil {
			return err
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		result, err := utils.ParseMarkdownWithMetadata[domain.RuleMetadata](body)
		if err != nil {
			return err
		}

		ruleItem := domain.NewRuleItem(slug, result.Content, result.FrontMatter)
		items = append(items, ruleItem)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return items, nil
}
