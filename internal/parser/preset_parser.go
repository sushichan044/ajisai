package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/frontmatter"
	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

// ParsePresetPackage scans the source directory for rules and prompts and returns a PresetPackage.
func ParsePresetPackage(config *domain.Config, presetName string) (*domain.PresetPackage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	if _, ok := config.Inputs[presetName]; !ok {
		return nil, fmt.Errorf("preset %s not found in config", presetName)
	}

	presetRootDir, err := resolvePresetRootDir(config, presetName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve preset root directory: %w", err)
	}

	var (
		prompts []*domain.PromptItem
		rules   []*domain.RuleItem
	)

	eg := new(errgroup.Group)

	eg.Go(func() error {
		parsedPrompts, innerErr := parsePrompts(presetRootDir)
		if innerErr != nil {
			return fmt.Errorf("failed to parse prompts: %w", innerErr)
		}

		prompts = parsedPrompts
		return nil
	})

	eg.Go(func() error {
		parsedRules, innerErr := parseRules(presetRootDir)
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
		Name:   presetName,
		Rule:   rules,
		Prompt: prompts,
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

		slug, err := getPromptSlug(rootDir, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var metadata domain.PromptMetadata

		rest, err := frontmatter.Parse(bytes.NewReader(content), &metadata)
		if err != nil {
			return err
		}

		ruleItem := domain.NewPromptItem(slug, string(rest), metadata)

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

		slug, err := getRuleSlug(rootDir, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var metadata domain.RuleMetadata

		rest, err := frontmatter.Parse(bytes.NewReader(content), &metadata)
		if err != nil {
			return err
		}

		ruleItem := domain.NewRuleItem(slug, string(rest), metadata)
		items = append(items, ruleItem)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return items, nil
}

func getRuleSlug(pkgRootDir string, fullPath string) (string, error) {
	relPath, err := filepath.Rel(pkgRootDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	return captureRuleSlug(relPath)
}

func getPromptSlug(pkgRootDir string, fullPath string) (string, error) {
	relPath, err := filepath.Rel(pkgRootDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	return capturePromptSlug(relPath)
}

func captureRuleSlug(relPath string) (string, error) {
	// e.g. rules/react/my-rule.md -> captures "react/my-rule"
	pattern := regexp.MustCompile(fmt.Sprintf("^%s/(.*)\\.%s$", domain.RulesPresetType, domain.RuleInternalExtension))
	if pattern == nil {
		return "", errors.New("failed to compile rule path regex")
	}

	matches := pattern.FindStringSubmatch(relPath)
	// e.g. rules/react/my-rule.md -> matches[0] = "rules/react/my-rule.md", matches[1] = "react/my-rule"
	expectMatches := 2
	if len(matches) < expectMatches {
		return "", fmt.Errorf("invalid rule path format: %s", relPath)
	}

	return matches[1], nil
}

func capturePromptSlug(relPath string) (string, error) {
	// e.g. prompts/react/my-prompt.md -> captures "react/my-prompt"
	pattern := regexp.MustCompile(
		fmt.Sprintf("^%s/(.*)\\.%s$", domain.PromptsPresetType, domain.PromptInternalExtension),
	)
	if pattern == nil {
		return "", errors.New("failed to compile prompt path regex")
	}

	matches := pattern.FindStringSubmatch(relPath)
	// e.g. prompts/react/my-prompt.md -> matches[0] = "prompts/react/my-prompt.md", matches[1] = "react/my-prompt"
	expectMatches := 2
	if len(matches) < expectMatches {
		return "", fmt.Errorf("invalid prompt path format: %s", relPath)
	}

	return matches[1], nil
}

func resolvePresetRootDir(config *domain.Config, presetName string) (string, error) {
	cacheDir, err := utils.ResolveAbsPath(config.Global.CacheDir)
	if err != nil {
		return "", err
	}

	inputConfig, ok := config.Inputs[presetName]
	if !ok {
		return "", fmt.Errorf("preset %s not found", presetName)
	}

	if _, isLocal := domain.GetInputSourceDetails[domain.LocalInputSourceDetails](inputConfig); isLocal {
		return filepath.Join(cacheDir, presetName), nil
	}

	if gitInput, isGit := domain.GetInputSourceDetails[domain.GitInputSourceDetails](inputConfig); isGit {
		if gitInput.SubDir != "" {
			return filepath.Join(cacheDir, presetName, gitInput.SubDir), nil
		}
		return filepath.Join(cacheDir, presetName), nil
	}

	return "", fmt.Errorf("invalid input source type: %s", inputConfig.Type)
}
