package integration

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
)

type agentSpecificationAdapter interface {
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

type integrationImpl struct {
	adapter agentSpecificationAdapter

	resolvedRulesRootDir   string
	resolvedPromptsRootDir string
}

func New(adapter agentSpecificationAdapter) (domain.AgentIntegration, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	resolvedRulesRootDir := filepath.Join(cwd, filepath.FromSlash(adapter.RulesDir()))
	resolvedPromptsRootDir := filepath.Join(cwd, filepath.FromSlash(adapter.PromptsDir()))

	return &integrationImpl{
		adapter:                adapter,
		resolvedRulesRootDir:   resolvedRulesRootDir,
		resolvedPromptsRootDir: resolvedPromptsRootDir,
	}, nil
}

func (repo *integrationImpl) WritePackage(namespace string, pkg *domain.AgentPresetPackage) error {
	eg := errgroup.Group{}

	// Create gitignore files for the namespace directories
	eg.Go(func() error {
		return repo.ensureGitignoreFiles(namespace)
	})

	for _, preset := range pkg.Presets {
		eg.Go(func() error {
			return repo.writePreset(namespace, preset)
		})
	}

	return eg.Wait()
}

//gocognit:ignore
func (repo *integrationImpl) writePreset(namespace string, preset *domain.AgentPreset) error {
	resolveRulePath := func(rule *domain.RuleItem) string {
		rulePath := rule.URI.GetInternalPath(repo.adapter.RuleExtension())
		return filepath.Join(repo.resolvedRulesRootDir, namespace, rulePath)
	}

	resolvePromptPath := func(prompt *domain.PromptItem) string {
		promptPath := prompt.URI.GetInternalPath(repo.adapter.PromptExtension())
		return filepath.Join(repo.resolvedPromptsRootDir, namespace, promptPath)
	}

	eg := errgroup.Group{}

	for _, rule := range preset.Rules {
		eg.Go(func() error {
			rulePath := resolveRulePath(rule)

			serialized, serializeErr := repo.adapter.SerializeRule(rule)
			if serializeErr != nil {
				return serializeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(rulePath)); dirErr != nil {
				return fmt.Errorf(
					"could not ensure dir for rule %s (URI: %s): %w",
					rulePath,
					rule.URI.String(),
					dirErr,
				)
			}

			return utils.AtomicWriteFile(rulePath, bytes.NewReader([]byte(serialized)))
		})
	}

	for _, prompt := range preset.Prompts {
		eg.Go(func() error {
			promptPath := resolvePromptPath(prompt)

			serialized, serializeErr := repo.adapter.SerializePrompt(prompt)
			if serializeErr != nil {
				return serializeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(promptPath)); dirErr != nil {
				return fmt.Errorf(
					"could not ensure dir for prompt %s (URI: %s): %w",
					promptPath,
					prompt.URI.String(),
					dirErr,
				)
			}

			return utils.AtomicWriteFile(promptPath, bytes.NewReader([]byte(serialized)))
		})
	}

	return eg.Wait()
}

// ensureGitignoreFiles creates .gitignore files in the namespace directories to ignore all contents.
func (repo *integrationImpl) ensureGitignoreFiles(namespace string) error {
	gitignoreContent := "*\n"

	ruleNamespaceDir := filepath.Join(repo.resolvedRulesRootDir, namespace)
	promptNamespaceDir := filepath.Join(repo.resolvedPromptsRootDir, namespace)

	eg := errgroup.Group{}

	// Create .gitignore for rules directory
	eg.Go(func() error {
		if dirErr := utils.EnsureDir(ruleNamespaceDir); dirErr != nil {
			return fmt.Errorf("could not ensure rules namespace dir %s: %w", ruleNamespaceDir, dirErr)
		}

		gitignorePath := filepath.Join(ruleNamespaceDir, ".gitignore")
		return utils.AtomicWriteFile(gitignorePath, bytes.NewReader([]byte(gitignoreContent)))
	})

	// Create .gitignore for prompts directory
	eg.Go(func() error {
		if dirErr := utils.EnsureDir(promptNamespaceDir); dirErr != nil {
			return fmt.Errorf("could not ensure prompts namespace dir %s: %w", promptNamespaceDir, dirErr)
		}

		gitignorePath := filepath.Join(promptNamespaceDir, ".gitignore")
		return utils.AtomicWriteFile(gitignorePath, bytes.NewReader([]byte(gitignoreContent)))
	})

	return eg.Wait()
}

func (repo *integrationImpl) Clean(namespace string) error {
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
