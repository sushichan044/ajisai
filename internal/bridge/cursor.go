package bridge

import (
	"fmt"
	"strings"

	"github.com/sushichan044/aisync/internal/domain"
)

type (
	CursorRule struct {
		Slug     string
		Content  string
		Metadata CursorRuleMetadata
	}

	CursorRuleMetadata struct {
		AlwaysApply bool   `yaml:"alwaysApply"`
		Description string `yaml:"description"`
		Globs       string `yaml:"globs"` // e.g. "**/*.{js,ts,jsx,tsx}"
	}

	CursorPrompt struct {
		Slug    string
		Content string
	}
)

type CursorBridge struct{}

func NewCursorBridge() domain.AgentBridge[CursorRule, CursorPrompt] {
	return &CursorBridge{}
}

func (bridge *CursorBridge) ToAgentRule(rule domain.RuleItem) (CursorRule, error) {
	switch rule.Metadata.Attach {
	case domain.AttachTypeAlways:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: true,
				Description: "",
				Globs:       "",
			},
		}, nil
	case domain.AttachTypeGlob:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: false,
				Description: "",
				Globs:       strings.Join(rule.Metadata.Glob, ","),
			},
		}, nil
	case domain.AttachTypeAgentRequested:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: false,
				Description: rule.Metadata.Description,
				Globs:       "",
			},
		}, nil
	case domain.AttachTypeManual:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: false,
				Description: "",
				Globs:       "",
			},
		}, nil
	default:
		return CursorRule{}, fmt.Errorf("unsupported rule attach type: %s", rule.Metadata.Attach)
	}
}

func (bridge *CursorBridge) FromAgentRule(rule CursorRule) (domain.RuleItem, error) {
	emptyGlobs := make([]string, 0)

	if rule.Metadata.AlwaysApply {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAlways,
				Glob:        emptyGlobs,
				Description: "",
			},
		), nil
	}

	if rule.Metadata.Globs != "" {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeGlob,
				Glob:        strings.Split(rule.Metadata.Globs, ","),
				Description: "",
			},
		), nil
	}

	if rule.Metadata.Description != "" {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAgentRequested,
				Description: rule.Metadata.Description,
				Glob:        emptyGlobs,
			},
		), nil
	}

	return *domain.NewRuleItem(
		rule.Slug,
		rule.Content,
		domain.RuleMetadata{
			Attach:      domain.AttachTypeManual,
			Description: "",
			Glob:        emptyGlobs,
		},
	), nil
}

func (bridge *CursorBridge) ToAgentPrompt(prompt domain.PromptItem) (CursorPrompt, error) {
	return CursorPrompt{
		Slug:    prompt.Slug,
		Content: prompt.Content,
	}, nil
}

func (bridge *CursorBridge) FromAgentPrompt(prompt CursorPrompt) (domain.PromptItem, error) {
	return *domain.NewPromptItem(
		prompt.Slug,
		prompt.Content,
		domain.PromptMetadata{},
	), nil
}

func (rule *CursorRule) String() (string, error) {
	// Cursor does not accept quoted front matters, so we need to write custom marshaler
	metaKeys := 3 // alwaysApply, description, globs
	metaContent := make([]string, 0, metaKeys)

	metaContent = append(metaContent, fmt.Sprintf("alwaysApply: %t", rule.Metadata.AlwaysApply))

	if rule.Metadata.Description == "" {
		metaContent = append(metaContent, "description:")
	} else {
		description := strings.TrimRight(rule.Metadata.Description, " ")
		metaContent = append(metaContent, fmt.Sprintf("description: %s", description))
	}

	if rule.Metadata.Globs == "" {
		metaContent = append(metaContent, "globs:")
	} else {
		globs := strings.TrimRight(rule.Metadata.Globs, " ")
		metaContent = append(metaContent, fmt.Sprintf("globs: %s", globs))
	}

	frontMatter := fmt.Sprintf("---\n%s\n---", strings.Join(metaContent, "\n"))

	// Special case: if the content is empty, we need to return just the front matter
	if rule.Content == "" {
		return frontMatter + "\n", nil
	}

	// Remove trailing newlines from the content, then add one newline at the end
	normalizedContent := strings.TrimRight(rule.Content, "\n")
	result := fmt.Sprintf("%s\n%s", frontMatter, normalizedContent)
	return result + "\n", nil
}

func (prompt *CursorPrompt) String() (string, error) {
	return prompt.Content, nil
}
