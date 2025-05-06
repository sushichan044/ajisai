package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

	presetRootDir, err := resolvePresetRootDir(config, presetName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve preset root directory: %w", err)
	}

	var (
		prompts []*domain.PresetItem
		rules   []*domain.PresetItem
	)

	eg := new(errgroup.Group)

	eg.Go(func() error {
		var err error
		prompts, err = parsePrompts(presetRootDir)
		if err != nil {
			return fmt.Errorf("failed to parse prompts: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		rules, err = parseRules(presetRootDir)
		if err != nil {
			return fmt.Errorf("failed to parse rules: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &domain.PresetPackage{
		InputKey: presetName,
		Items:    append(prompts, rules...),
	}, nil
}

func parsePrompts(rootDir string) ([]*domain.PresetItem, error) {
	promptRootDir := filepath.Join(rootDir, "prompts")
	items := []*domain.PresetItem{}

	if exists, err := utils.IsDirExists(promptRootDir); err != nil {
		return nil, fmt.Errorf("failed to check if prompt directory exists: %w", err)
	} else if !exists {
		return items, nil
	}

	walkErr := filepath.WalkDir(promptRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		fileName := d.Name()
		ext := filepath.Ext(fileName)
		if ext != ".md" {
			return nil
		}

		slug := strings.TrimSuffix(fileName, ext)

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var metadata domain.PromptMetadata

		rest, err := frontmatter.Parse(bytes.NewReader(content), &metadata)
		if err != nil {
			return err
		}

		items = append(items, &domain.PresetItem{
			Name:         slug,
			Content:      string(rest),
			Type:         domain.PromptPresetType,
			Metadata:     metadata,
			RelativePath: relPath,
		})
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return items, nil
}

func parseRules(rootDir string) ([]*domain.PresetItem, error) {
	ruleRootDir := filepath.Join(rootDir, "rules")
	items := []*domain.PresetItem{}

	if exists, err := utils.IsDirExists(ruleRootDir); err != nil {
		return nil, fmt.Errorf("failed to check if rule directory exists: %w", err)
	} else if !exists {
		return items, nil
	}

	walkErr := filepath.WalkDir(ruleRootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		fileName := d.Name()
		ext := filepath.Ext(fileName)
		if ext != ".md" {
			return nil
		}

		slug := strings.TrimSuffix(fileName, ext)

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var metadata domain.RuleMetadata

		rest, err := frontmatter.Parse(bytes.NewReader(content), &metadata)
		if err != nil {
			return err
		}

		items = append(items, &domain.PresetItem{
			Name:         slug,
			Content:      string(rest),
			Type:         domain.RulePresetType,
			Metadata:     metadata,
			RelativePath: relPath,
		})
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return items, nil
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
