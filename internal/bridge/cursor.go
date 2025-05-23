package bridge

import (
	"strconv"
	"strings"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
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
		Globs       string `yaml:"globs"`
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
				Globs:       strings.Join(rule.Metadata.Globs, ","),
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
	}

	// Fallback as manual rule.
	return CursorRule{
		Slug:    rule.Slug,
		Content: rule.Content,
		Metadata: CursorRuleMetadata{
			AlwaysApply: false,
			Description: "",
			Globs:       "",
		},
	}, nil
}

func (bridge *CursorBridge) FromAgentRule(rule CursorRule) (domain.RuleItem, error) {
	emptyGlobs := make([]string, 0)

	if rule.Metadata.AlwaysApply {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAlways,
				Globs:       emptyGlobs,
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
				Globs:       strings.Split(rule.Metadata.Globs, ","),
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
				Globs:       emptyGlobs,
			},
		), nil
	}

	return *domain.NewRuleItem(
		rule.Slug,
		rule.Content,
		domain.RuleMetadata{
			Attach:      domain.AttachTypeManual,
			Description: "",
			Globs:       emptyGlobs,
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

func (bridge *CursorBridge) SerializeAgentRule(rule CursorRule) (string, error) {
	// Cursor does not accept quoted front matters, so we need to write custom marshaler
	metaKeys := 3 // alwaysApply, description, globs
	metaContent := make([]string, 0, metaKeys)

	metaContent = append(metaContent, "alwaysApply: "+strconv.FormatBool(rule.Metadata.AlwaysApply))

	if desc := strings.TrimSpace(rule.Metadata.Description); desc != "" {
		metaContent = append(metaContent, "description: "+desc)
	} else {
		metaContent = append(metaContent, "description:")
	}

	if globs := strings.TrimSpace(rule.Metadata.Globs); globs != "" {
		metaContent = append(metaContent, "globs: "+globs)
	} else {
		metaContent = append(metaContent, "globs:")
	}

	return strings.TrimRight("---\n"+strings.Join(metaContent, "\n")+"\n---\n"+rule.Content, "\n") + "\n", nil
}

func (bridge *CursorBridge) DeserializeAgentRule(slug string, ruleBody string) (CursorRule, error) {
	lines := strings.Split(ruleBody, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "globs: ") {
			// we need to add quotes around the glob patterns to avoid parsing errors
			lines[i] = "globs: " + strconv.Quote(strings.TrimSpace(line[7:]))
		}
	}
	ruleBody = strings.Join(lines, "\n")

	result, err := utils.ParseMarkdownWithMetadata[CursorRuleMetadata]([]byte(ruleBody))
	if err != nil {
		return CursorRule{}, err
	}

	return CursorRule{
		Slug:     slug,
		Content:  result.Content,
		Metadata: result.FrontMatter,
	}, nil
}

func (bridge *CursorBridge) SerializeAgentPrompt(prompt CursorPrompt) (string, error) {
	return prompt.Content, nil
}

func (bridge *CursorBridge) DeserializeAgentPrompt(slug string, promptBody string) (CursorPrompt, error) {
	return CursorPrompt{
		Slug:    slug,
		Content: promptBody,
	}, nil
}
