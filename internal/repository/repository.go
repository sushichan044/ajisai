package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type (
	AgentFileAdapter interface {
		/*
			Returns the extension for rules. (e.g. `.instructions.md`)
		*/
		RuleExtension() string

		/*
			Returns the extension for prompts. (e.g. `.prompt.md`)
		*/
		PromptExtension() string

		/*
			Returns the directory path for rules. (e.g. `.github/instructions`)

		*/
		RulesDir() string

		/*
			Returns the directory path for prompts. (e.g. `.github/prompts`)
		*/
		PromptsDir() string

		SerializeRule(rule *domain.RuleItem) (string, error)

		SerializePrompt(prompt *domain.PromptItem) (string, error)
	}
)

type repositoryImpl struct {
	adapter AgentFileAdapter

	resolvedRulesRootDir   string
	resolvedPromptsRootDir string
}

func NewPresetRepository(adapter AgentFileAdapter) (domain.PresetRepository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	resolvedRulesRootDir := filepath.Join(cwd, filepath.FromSlash(adapter.RulesDir()))
	resolvedPromptsRootDir := filepath.Join(cwd, filepath.FromSlash(adapter.PromptsDir()))

	return &repositoryImpl{
		adapter:                adapter,
		resolvedRulesRootDir:   resolvedRulesRootDir,
		resolvedPromptsRootDir: resolvedPromptsRootDir,
	}, nil
}

func (repo *repositoryImpl) WritePackage(namespace string, pkg *domain.AgentPresetPackage) error {
	eg := errgroup.Group{}

	for _, preset := range pkg.Presets {
		eg.Go(func() error {
			return repo.writePreset(namespace, pkg.PackageName, preset)
		})
	}

	return eg.Wait()
}

//gocognit:ignore
func (repo *repositoryImpl) writePreset(namespace string, packageName string, preset *domain.AgentPreset) error {
	resolveRulePath := func(rule *domain.RuleItem) (string, error) {
		rulePath, err := rule.GetInternalPath(packageName, preset.Name, repo.adapter.RuleExtension())
		if err != nil {
			return "", err
		}

		return filepath.Join(repo.resolvedRulesRootDir, namespace, rulePath), nil
	}

	resolvePromptPath := func(prompt *domain.PromptItem) (string, error) {
		promptPath, err := prompt.GetInternalPath(packageName, preset.Name, repo.adapter.PromptExtension())
		if err != nil {
			return "", err
		}

		return filepath.Join(repo.resolvedPromptsRootDir, namespace, promptPath), nil
	}

	eg := errgroup.Group{}

	for _, rule := range preset.Rules {
		eg.Go(func() error {
			rulePath, err := resolveRulePath(rule)
			if err != nil {
				return err
			}

			serialized, serializeErr := repo.adapter.SerializeRule(rule)
			if serializeErr != nil {
				return serializeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(rulePath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for rule %s: %w", rulePath, dirErr)
			}

			return os.WriteFile(rulePath, []byte(serialized), 0600)
		})
	}

	for _, prompt := range preset.Prompts {
		eg.Go(func() error {
			promptPath, err := resolvePromptPath(prompt)
			if err != nil {
				return err
			}

			serialized, serializeErr := repo.adapter.SerializePrompt(prompt)
			if serializeErr != nil {
				return serializeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(promptPath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for prompt %s: %w", promptPath, dirErr)
			}

			return os.WriteFile(promptPath, []byte(serialized), 0600)
		})
	}

	return eg.Wait()
}

func (repo *repositoryImpl) Clean(namespace string) error {
	ruleDir := filepath.Join(repo.resolvedRulesRootDir, namespace)
	promptDir := filepath.Join(repo.resolvedPromptsRootDir, namespace)

	eg := errgroup.Group{}

	eg.Go(func() error {
		return os.RemoveAll(ruleDir)
	})

	eg.Go(func() error {
		return os.RemoveAll(promptDir)
	})

	return eg.Wait()
}
