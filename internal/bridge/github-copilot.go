package bridge

import (
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
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
			Slug:    rule.URI.Path,
			Content: rule.Content,
			Metadata: GitHubCopilotInstructionMetadata{
				ApplyTo: GitHubCopilotApplyToAllPrimary,
			},
		}, nil
	case domain.AttachTypeGlob:
		return GitHubCopilotInstruction{
			Slug:    rule.URI.Path,
			Content: rule.Content,
			Metadata: GitHubCopilotInstructionMetadata{
				ApplyTo: strings.Join(rule.Metadata.Globs, ","),
			},
		}, nil
	case domain.AttachTypeAgentRequested, domain.AttachTypeManual:
		return GitHubCopilotInstruction{
			Slug:     rule.URI.Path,
			Content:  rule.Content,
			Metadata: GitHubCopilotInstructionMetadata{},
		}, nil
	}

	// Fallback as manual rule.
	return GitHubCopilotInstruction{
		Slug:     rule.URI.Path,
		Content:  rule.Content,
		Metadata: GitHubCopilotInstructionMetadata{},
	}, nil
}

func (bridge *GitHubCopilotBridge) FromAgentRule(rule GitHubCopilotInstruction) (domain.RuleItem, error) {
	emptyGlobs := make([]string, 0)

	globs := utils.RemoveZeroValues(strings.Split(rule.Metadata.ApplyTo, ","))
	alwaysApplied := utils.ContainsAny(globs, GitHubCopilotInstructionApplyToAll)

	if alwaysApplied {
		return *domain.NewRuleItem(
			"", // packageName - placeholder, bridge doesn't have this context
			"", // presetName - placeholder, bridge doesn't have this context
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach: domain.AttachTypeAlways,
				Globs:  emptyGlobs,
			},
		), nil
	}

	if len(globs) > 0 {
		return *domain.NewRuleItem(
			"", // packageName - placeholder, bridge doesn't have this context
			"", // presetName - placeholder, bridge doesn't have this context
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach: domain.AttachTypeGlob,
				Globs:  globs,
			},
		), nil
	}

	return *domain.NewRuleItem(
		"", // packageName - placeholder, bridge doesn't have this context
		"", // presetName - placeholder, bridge doesn't have this context
		rule.Slug,
		rule.Content,
		domain.RuleMetadata{
			Attach: domain.AttachTypeManual,
			Globs:  emptyGlobs,
		},
	), nil
}

func (bridge *GitHubCopilotBridge) ToAgentPrompt(prompt domain.PromptItem) (GitHubCopilotPrompt, error) {
	return GitHubCopilotPrompt{
		Slug:    prompt.URI.Path,
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
		"", // packageName - placeholder, bridge doesn't have this context
		"", // presetName - placeholder, bridge doesn't have this context
		prompt.Slug,
		prompt.Content,
		domain.PromptMetadata{
			Description: prompt.Metadata.Description,
		},
	), nil
}

func (bridge *GitHubCopilotBridge) SerializeAgentRule(rule GitHubCopilotInstruction) (string, error) {
	frontMatterBytes, err := yaml.Marshal(rule.Metadata)
	if err != nil {
		return "", err
	}

	metadata := string(frontMatterBytes)
	tidyContent := strings.TrimRight(rule.Content, "\n")

	if strings.TrimSpace(metadata) == "{}" {
		// If the metadata is empty, return the content only.
		return tidyContent, nil
	}

	return "---\n" + metadata + "---\n" + tidyContent + "\n", nil
}

func (bridge *GitHubCopilotBridge) DeserializeAgentRule(
	slug string,
	ruleBody string,
) (GitHubCopilotInstruction, error) {
	result, err := utils.ParseMarkdownWithMetadata[GitHubCopilotInstructionMetadata]([]byte(ruleBody))
	if err != nil {
		return GitHubCopilotInstruction{}, err
	}

	return GitHubCopilotInstruction{
		Slug:     slug,
		Content:  result.Content,
		Metadata: result.FrontMatter,
	}, nil
}

func (bridge *GitHubCopilotBridge) SerializeAgentPrompt(prompt GitHubCopilotPrompt) (string, error) {
	frontMatterBytes, err := yaml.Marshal(prompt.Metadata)
	if err != nil {
		return "", err
	}

	metadata := string(frontMatterBytes)

	if strings.TrimSpace(metadata) == "{}" {
		// If the metadata is empty, return the content only.
		return strings.TrimRight(prompt.Content, "\n") + "\n", nil
	}
	return strings.TrimRight("---\n"+metadata+"---\n"+prompt.Content, "\n") + "\n", nil
}

func (bridge *GitHubCopilotBridge) DeserializeAgentPrompt(
	slug string,
	promptBody string,
) (GitHubCopilotPrompt, error) {
	result, err := utils.ParseMarkdownWithMetadata[GitHubCopilotPromptMetadata]([]byte(promptBody))
	if err != nil {
		return GitHubCopilotPrompt{}, err
	}

	return GitHubCopilotPrompt{
		Slug:     slug,
		Content:  result.Content,
		Metadata: result.FrontMatter,
	}, nil
}
