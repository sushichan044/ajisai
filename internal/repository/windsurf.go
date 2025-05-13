package repository

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type WindsurfRepository struct {
	rulesRootDir   string
	promptsRootDir string

	bridge domain.AgentBridge[bridge.WindsurfRule, bridge.WindsurfPrompt]
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
		bridge:         bridge.NewWindsurfBridge(),
	}, nil
}

// NewWindsurfRepositoryWithPaths creates a new WindsurfRepository with custom paths.
// This is mainly used for testing.
func NewWindsurfRepositoryWithPaths(rulesDir, promptsDir string) (*WindsurfRepository, error) {
	return &WindsurfRepository{
		rulesRootDir:   rulesDir,
		promptsRootDir: promptsDir,
		bridge:         bridge.NewWindsurfBridge(),
	}, nil
}

//gocognit:ignore
func (repo *WindsurfRepository) WritePackage(namespace string, pkg domain.PresetPackage) error {
	resolveRulePath := func(rule *domain.RuleItem) (string, error) {
		rulePath, err := rule.GetInternalPath(namespace, pkg.Name, WindsurfRuleExtension)
		if err != nil {
			return "", err
		}

		return filepath.Join(repo.rulesRootDir, rulePath), nil
	}

	resolvePromptPath := func(prompt *domain.PromptItem) (string, error) {
		promptPath, err := prompt.GetInternalPath(namespace, pkg.Name, WindsurfPromptExtension)
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

			ruleItem, err := repo.bridge.ToAgentRule(*rule)
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

			prompt, promptConversionErr := repo.bridge.ToAgentPrompt(*prompt)
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

func (repo *WindsurfRepository) ReadPackage(_ string) (domain.PresetPackage, error) {
	return domain.PresetPackage{}, nil
}

func (repo *WindsurfRepository) ReadRules(namespace string) ([]*domain.RuleItem, error) {
	ruleDir := filepath.Join(repo.rulesRootDir, namespace)
	rules := []*domain.RuleItem{}

	walkErr := filepath.WalkDir(ruleDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(d.Name()) != "."+WindsurfRuleExtension {
			return nil
		}

		slug, slugErr := utils.GetSlugFromBaseDir(ruleDir, path)
		if slugErr != nil {
			return slugErr
		}

		rawBody, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		result, parseErr := utils.ParseMarkdownWithMetadata[bridge.WindsurfRuleMetadata](rawBody)
		if parseErr != nil {
			return parseErr
		}

		ruleItem, bridgeErr := repo.bridge.FromAgentRule(bridge.WindsurfRule{
			Slug:     slug,
			Content:  result.Content,
			Metadata: result.FrontMatter,
		})
		if bridgeErr != nil {
			return bridgeErr
		}

		rules = append(rules, &ruleItem)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	return rules, nil
}

func (repo *WindsurfRepository) Clean(namespace string) error {
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
