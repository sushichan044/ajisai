package bridge

import (
	"fmt"
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

type (
	GitHubCopilotChatMode string

	GitHubCopilotInstruction struct {
		Slug     string
		Content  string
		Metadata GitHubCopilotInstructionMetadata
	}

	GitHubCopilotInstructionMetadata struct {
		ApplyTo string `yaml:"applyTo,omitempty"`
	}

	GitHubCopilotPrompt struct {
		Slug     string
		Content  string
		Metadata GitHubCopilotPromptMetadata
	}

	GitHubCopilotPromptMetadata struct {
		Description string `yaml:"description,omitempty"`
		// The chat mode to use when running the prompt: ask, edit, or agent (default).
		Mode  GitHubCopilotChatMode `yaml:"mode,omitempty"`
		Tools []string              `yaml:"tools,omitempty"`
	}
)

const (
	GitHubCopilotInstructionModeAgent GitHubCopilotChatMode = "agent"
	GitHubCopilotInstructionModeAsk   GitHubCopilotChatMode = "ask"
	GitHubCopilotInstructionModeEdit  GitHubCopilotChatMode = "edit"

	// GitHubCopilotApplyToAllPrimary and GitHubCopilotApplyToAllSecondary are treated as the same internally.

	// GitHubCopilotApplyToAllPrimary is the primary glob pattern for applying to all files.
	GitHubCopilotApplyToAllPrimary string = "**"
	// GitHubCopilotApplyToAllSecondary is the secondary glob pattern for applying to all files.
	GitHubCopilotApplyToAllSecondary string = "**/*"
)

var (
	// GitHubCopilotInstructionApplyToAll is special glob pattern
	// that means the instruction should be applied to all files.
	GitHubCopilotInstructionApplyToAll = []string{
		GitHubCopilotApplyToAllPrimary,
		GitHubCopilotApplyToAllSecondary,
	}
)

type GitHubCopilotBridge struct{}

func NewGitHubCopilotBridge() domain.AgentBridge[GitHubCopilotInstruction, GitHubCopilotPrompt] {
	return &GitHubCopilotBridge{}
}

func (bridge *GitHubCopilotBridge) ToAgentRule(rule domain.RuleItem) (GitHubCopilotInstruction, error) {
	switch rule.Metadata.Attach {
	case domain.AttachTypeAlways:
		return GitHubCopilotInstruction{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: GitHubCopilotInstructionMetadata{
				ApplyTo: GitHubCopilotApplyToAllPrimary,
			},
		}, nil
	case domain.AttachTypeGlob:
		return GitHubCopilotInstruction{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: GitHubCopilotInstructionMetadata{
				ApplyTo: strings.Join(rule.Metadata.Glob, ","),
			},
		}, nil
	case domain.AttachTypeAgentRequested, domain.AttachTypeManual:
		return GitHubCopilotInstruction{
			Slug:     rule.Slug,
			Content:  rule.Content,
			Metadata: GitHubCopilotInstructionMetadata{},
		}, nil
	default:
		return GitHubCopilotInstruction{}, fmt.Errorf("unsupported rule attach type: %s", rule.Metadata.Attach)
	}
}

func (bridge *GitHubCopilotBridge) FromAgentRule(rule GitHubCopilotInstruction) (domain.RuleItem, error) {
	emptyGlobs := make([]string, 0)

	globs := utils.RemoveZeroValues(strings.Split(rule.Metadata.ApplyTo, ","))
	alwaysApplied := utils.ContainsAny(globs, GitHubCopilotInstructionApplyToAll)

	if alwaysApplied {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach: domain.AttachTypeAlways,
				Glob:   emptyGlobs,
			},
		), nil
	}

	if len(globs) > 0 {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach: domain.AttachTypeGlob,
				Glob:   globs,
			},
		), nil
	}

	return *domain.NewRuleItem(
		rule.Slug,
		rule.Content,
		domain.RuleMetadata{
			Attach: domain.AttachTypeManual,
			Glob:   emptyGlobs,
		},
	), nil
}

func (bridge *GitHubCopilotBridge) ToAgentPrompt(prompt domain.PromptItem) (GitHubCopilotPrompt, error) {
	return GitHubCopilotPrompt{
		Slug:    prompt.Slug,
		Content: prompt.Content,
		Metadata: GitHubCopilotPromptMetadata{
			Description: prompt.Metadata.Description,
			// TODO: Add support Mode, Tools.
			Mode:  GitHubCopilotInstructionModeAgent,
			Tools: []string{},
		},
	}, nil
}

func (bridge *GitHubCopilotBridge) FromAgentPrompt(prompt GitHubCopilotPrompt) (domain.PromptItem, error) {
	return *domain.NewPromptItem(
		prompt.Slug,
		prompt.Content,
		domain.PromptMetadata{
			Description: prompt.Metadata.Description,
		},
	), nil
}

func (instruction *GitHubCopilotInstruction) String() (string, error) {
	frontMatterBytes, err := yaml.MarshalWithOptions(instruction.Metadata)
	if err != nil {
		return "", err
	}

	metadata := string(frontMatterBytes)

	if strings.TrimSpace(metadata) == "{}" {
		// If the metadata is empty, return the content only.
		return instruction.Content, nil
	}

	return fmt.Sprintf("---\n%s---\n\n%s", metadata, instruction.Content), nil
}

func (prompt *GitHubCopilotPrompt) String() (string, error) {
	frontMatterBytes, err := yaml.MarshalWithOptions(prompt.Metadata)
	if err != nil {
		return "", err
	}

	metadata := string(frontMatterBytes)

	if strings.TrimSpace(metadata) == "{}" {
		// If the metadata is empty, return the content only.
		return prompt.Content, nil
	}
	return fmt.Sprintf("---\n%s---\n\n%s", metadata, prompt.Content), nil
}
