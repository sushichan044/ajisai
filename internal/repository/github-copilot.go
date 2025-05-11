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

type GitHubCopilotRepository struct {
	instructionsRootDir string
	promptsRootDir      string
}

func NewGitHubCopilotRepository() (domain.PresetRepository, error) {
	cwd, wdErr := os.Getwd()
	if wdErr != nil {
		return nil, wdErr
	}

	return &GitHubCopilotRepository{
		instructionsRootDir: filepath.Join(cwd, ".github", "instructions"),
		promptsRootDir:      filepath.Join(cwd, ".github", "prompts"),
	}, nil
}

func NewGitHubCopilotRepositoryWithPaths(instructionsDir, promptsDir string) (*GitHubCopilotRepository, error) {
	return &GitHubCopilotRepository{
		instructionsRootDir: instructionsDir,
		promptsRootDir:      promptsDir,
	}, nil
}

const (
	GitHubCopilotInstructionExtension = "instructions.md"
	GitHubCopilotPromptExtension      = "prompt.md"
)

//gocognit:ignore
func (repository *GitHubCopilotRepository) WritePackage(namespace string, pkg domain.PresetPackage) error {
	bridge := bridge.NewGitHubCopilotBridge()

	resolveInstructionPath := func(instruction *domain.RuleItem) (string, error) {
		instructionPath, innerErr := instruction.GetInternalPath(
			namespace,
			pkg.Name,
			GitHubCopilotInstructionExtension,
		)
		if innerErr != nil {
			return "", innerErr
		}

		return filepath.Join(repository.instructionsRootDir, instructionPath), nil
	}

	resolvePromptPath := func(prompt *domain.PromptItem) (string, error) {
		promptPath, innerErr := prompt.GetInternalPath(
			namespace,
			pkg.Name,
			GitHubCopilotPromptExtension,
		)
		if innerErr != nil {
			return "", innerErr
		}

		return filepath.Join(repository.promptsRootDir, promptPath), nil
	}

	eg := errgroup.Group{}

	for _, rule := range pkg.Rule {
		eg.Go(func() error {
			instructionPath, pathErr := resolveInstructionPath(rule)
			if pathErr != nil {
				return pathErr
			}

			instruction, bridgeErr := bridge.ToAgentRule(*rule)
			if bridgeErr != nil {
				return bridgeErr
			}

			instructionStr, serializeErr := instruction.String()
			if serializeErr != nil {
				return serializeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(instructionPath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for instruction %s: %w", instructionPath, dirErr)
			}

			return os.WriteFile(instructionPath, []byte(instructionStr), 0600)
		})
	}

	for _, prompt := range pkg.Prompt {
		eg.Go(func() error {
			promptPath, pathErr := resolvePromptPath(prompt)
			if pathErr != nil {
				return pathErr
			}

			prompt, bridgeErr := bridge.ToAgentPrompt(*prompt)
			if bridgeErr != nil {
				return bridgeErr
			}

			promptStr, serializeErr := prompt.String()
			if serializeErr != nil {
				return serializeErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(promptPath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for prompt %s: %w", promptPath, dirErr)
			}

			return os.WriteFile(promptPath, []byte(promptStr), 0600)
		})
	}

	return eg.Wait()
}

func (repository *GitHubCopilotRepository) ReadPackage(_ string) (domain.PresetPackage, error) {
	return domain.PresetPackage{}, nil
}

func (repository *GitHubCopilotRepository) Clean(namespace string) error {
	instructionDir := filepath.Join(repository.instructionsRootDir, namespace)
	promptDir := filepath.Join(repository.promptsRootDir, namespace)

	eg := errgroup.Group{}

	eg.Go(func() error {
		return os.RemoveAll(instructionDir)
	})

	eg.Go(func() error {
		return os.RemoveAll(promptDir)
	})

	return eg.Wait()
}
