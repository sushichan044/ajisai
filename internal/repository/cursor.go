package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type CursorRepository struct {
	rulesRootDir   string
	promptsRootDir string

	bridge domain.AgentBridge[bridge.CursorRule, bridge.CursorPrompt]
}

const (
	CursorRuleExtension   = "mdc"
	CursorPromptExtension = "md"
)

func NewCursorRepository() (domain.PresetRepository, error) {
	cwd, wdErr := os.Getwd()
	if wdErr != nil {
		return nil, wdErr
	}

	return &CursorRepository{
		rulesRootDir:   filepath.Join(cwd, ".cursor", "rules"),
		promptsRootDir: filepath.Join(cwd, ".cursor", "prompts"),
		bridge:         bridge.NewCursorBridge(),
	}, nil
}

// NewCursorRepositoryWithPaths creates a new CursorRepository with custom paths.
// This is mainly used for testing.
func NewCursorRepositoryWithPaths(rulesDir, promptsDir string) (*CursorRepository, error) {
	return &CursorRepository{
		bridge:         bridge.NewCursorBridge(),
		rulesRootDir:   rulesDir,
		promptsRootDir: promptsDir,
	}, nil
}

//gocognit:ignore
func (repo *CursorRepository) WritePackage(namespace string, pkg domain.PresetPackage) error {
	resolveRulePath := func(rule *domain.RuleItem) (string, error) {
		rulePath, err := rule.GetInternalPath(namespace, pkg.Name, CursorRuleExtension)
		if err != nil {
			return "", err
		}

		return filepath.Join(repo.rulesRootDir, rulePath), nil
	}

	resolvePromptPath := func(prompt *domain.PromptItem) (string, error) {
		promptPath, err := prompt.GetInternalPath(namespace, pkg.Name, CursorPromptExtension)
		if err != nil {
			return "", err
		}

		return filepath.Join(repo.promptsRootDir, promptPath), nil
	}

	eg := errgroup.Group{}

	for _, rule := range pkg.Rules {
		eg.Go(func() error {
			rulePath, err := resolveRulePath(rule)
			if err != nil {
				return err
			}

			cursorRule, ruleConversionErr := repo.bridge.ToAgentRule(*rule)
			if ruleConversionErr != nil {
				return ruleConversionErr
			}

			cursorRuleStr, ruleStrErr := cursorRule.String()
			if ruleStrErr != nil {
				return ruleStrErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(rulePath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for rule %s: %w", rulePath, dirErr)
			}

			return os.WriteFile(rulePath, []byte(cursorRuleStr), 0600)
		})
	}

	for _, prompt := range pkg.Prompts {
		eg.Go(func() error {
			promptPath, err := resolvePromptPath(prompt)
			if err != nil {
				return err
			}

			cursorPrompt, promptConversionErr := repo.bridge.ToAgentPrompt(*prompt)
			if promptConversionErr != nil {
				return promptConversionErr
			}

			cursorPromptStr, promptStrErr := cursorPrompt.String()
			if promptStrErr != nil {
				return promptStrErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(promptPath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for prompt %s: %w", promptPath, dirErr)
			}

			return os.WriteFile(promptPath, []byte(cursorPromptStr), 0600)
		})
	}

	return eg.Wait()
}

func (repo *CursorRepository) ReadPackage(_ string) (domain.PresetPackage, error) {
	return domain.PresetPackage{}, nil
}

func (repo *CursorRepository) Clean(namespace string) error {
	ruleDir := filepath.Join(repo.rulesRootDir, namespace)
	promptDir := filepath.Join(repo.promptsRootDir, namespace)

	eg := errgroup.Group{}

	eg.Go(func() error {
		return os.RemoveAll(ruleDir)
	})

	eg.Go(func() error {
		return os.RemoveAll(promptDir)
	})

	return eg.Wait()
}
