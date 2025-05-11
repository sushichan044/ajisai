package bridge

import (
	"fmt"
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/sushichan044/aisync/internal/domain"
)

type CursorRule struct {
	Slug     string
	Content  string
	Metadata CursorRuleMetadata
}

type CursorRuleMetadata struct {
	AlwaysApply bool   `yaml:"alwaysApply"`
	Description string `yaml:"description"`
	Globs       string `yaml:"globs"` // e.g. "**/*.{js,ts,jsx,tsx}"
}

type CursorPrompt struct {
	Slug    string
	Content string
}

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
	frontMatterBytes, err := yaml.Marshal(rule.Metadata)
	if err != nil {
		return "", err
	}

	fmStr := string(frontMatterBytes)
	var resultLines []string
	lines := strings.SplitSeq(strings.TrimRight(fmStr, "\n"), "\n")

	// Cursor only accepts non-standard YAML formatting, so we need to write special encoding logic.
	for line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Handle empty description: `description: ""` -> `description:`
		if trimmedLine == `description: ""` {
			resultLines = append(resultLines, "description:")
			continue
		}

		// Handle empty globs: `globs: ""` -> `globs:`
		if trimmedLine == `globs: ""` {
			resultLines = append(resultLines, "globs:")
			continue
		}

		// Handle non-empty description: `description: "..."` -> `description: '... '`
		// It should be single-quoted, single-line, with a trailing space inside the quotes.
		if strings.HasPrefix(line, "description: ") && rule.Metadata.Description != "" {
			originalDesc := rule.Metadata.Description
			// Replace newlines in original description with spaces to ensure it's single line
			singleLineDesc := strings.ReplaceAll(originalDesc, "\n", " ")
			// Trim any existing trailing spaces from the original description before adding our specific one
			singleLineDesc = strings.TrimRight(singleLineDesc, " ")
			resultLines = append(resultLines, fmt.Sprintf("description: '%s '", singleLineDesc))
			continue
		}

		// Handle non-empty globs: `globs: "content"` -> `globs: content` (remove quotes)
		if strings.HasPrefix(line, "globs: ") && rule.Metadata.Globs != "" {
			content := strings.TrimPrefix(line, "globs: ")
			content = strings.Trim(content, `"`) // Remove surrounding double quotes from default marshalling
			resultLines = append(resultLines, "globs: "+content)
			continue
		}

		resultLines = append(resultLines, line)
	}

	fmStr = strings.Join(resultLines, "\n") + "\n"

	var normalizedContent string
	if rule.Content != "" {
		normalizedContent = strings.TrimRight(rule.Content, "\n") + "\n"
	}

	return fmt.Sprintf("---\n%s---\n%s", fmStr, normalizedContent), nil
}

func (prompt *CursorPrompt) String() (string, error) {
	return prompt.Content, nil
}
