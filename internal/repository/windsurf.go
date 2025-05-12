package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/aisync/internal/bridge"
	"github.com/sushichan044/aisync/internal/domain"
	"github.com/sushichan044/aisync/utils"
)

type WindsurfRepository struct {
	rulesRootDir   string
	promptsRootDir string
}

const (
	WindsurfRuleExtension   = "md"
	WindsurfPromptExtension = "md"
)

func NewWindsurfRepository() (domain.PresetRepository, error) {
	cwd, wdErr := os.Getwd()
	if wdErr != nil {
		return nil, wdErr
	}

	return &WindsurfRepository{
		rulesRootDir:   filepath.Join(cwd, ".windsurf", "rules"),
		promptsRootDir: filepath.Join(cwd, ".windsurf", "prompts"),
	}, nil
}

// NewWindsurfRepositoryWithPaths creates a new WindsurfRepository with custom paths.
// This is mainly used for testing.
func NewWindsurfRepositoryWithPaths(rulesDir, promptsDir string) (*WindsurfRepository, error) {
	return &WindsurfRepository{
		rulesRootDir:   rulesDir,
		promptsRootDir: promptsDir,
	}, nil
}

//gocognit:ignore
func (repository *WindsurfRepository) WritePackage(namespace string, pkg domain.PresetPackage) error {
	bridge := bridge.NewWindsurfBridge()

	resolveRulePath := func(rule *domain.RuleItem) (string, error) {
		rulePath, err := rule.GetInternalPath(namespace, pkg.Name, WindsurfRuleExtension)
		if err != nil {
			return "", err
		}

		return filepath.Join(repository.rulesRootDir, rulePath), nil
	}

	resolvePromptPath := func(prompt *domain.PromptItem) (string, error) {
		promptPath, err := prompt.GetInternalPath(namespace, pkg.Name, WindsurfPromptExtension)
		if err != nil {
			return "", err
		}

		return filepath.Join(repository.promptsRootDir, promptPath), nil
	}

	eg := errgroup.Group{}
	for _, rule := range pkg.Rules {
		eg.Go(func() error {
			rulePath, err := resolveRulePath(rule)
			if err != nil {
				return err
			}

			ruleItem, err := bridge.ToAgentRule(*rule)
			if err != nil {
				return err
			}

			ruleStr, encodeErr := ruleItem.String()
			if encodeErr != nil {
				return encodeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(rulePath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for rule %s: %w", rulePath, dirErr)
			}

			return os.WriteFile(rulePath, []byte(ruleStr), 0600)
		})
	}

	for _, prompt := range pkg.Prompts {
		eg.Go(func() error {
			promptPath, err := resolvePromptPath(prompt)
			if err != nil {
				return err
			}

			prompt, promptConversionErr := bridge.ToAgentPrompt(*prompt)
			if promptConversionErr != nil {
				return promptConversionErr
			}

			promptStr, encodeErr := prompt.String()
			if encodeErr != nil {
				return encodeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(promptPath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for prompt %s: %w", promptPath, dirErr)
			}

			return os.WriteFile(promptPath, []byte(promptStr), 0600)
		})
	}

	return eg.Wait()
}

func (repository *WindsurfRepository) ReadPackage(_ string) (domain.PresetPackage, error) {
	return domain.PresetPackage{}, nil
}

func (repository *WindsurfRepository) Clean(namespace string) error {
	ruleDir := filepath.Join(repository.rulesRootDir, namespace)
	promptDir := filepath.Join(repository.promptsRootDir, namespace)

	eg := errgroup.Group{}

	eg.Go(func() error {
		return os.RemoveAll(ruleDir)
	})
	eg.Go(func() error {
		return os.RemoveAll(promptDir)
	})

	return eg.Wait()
}
