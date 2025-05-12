package bridge

import (
	"fmt"
	"strings"

	"github.com/sushichan044/aisync/internal/domain"
)

type (
	WindsurfRule struct {
		Slug     string
		Content  string
		Metadata WindsurfRuleMetadata
	}

	WindsurfRuleMetadata struct {
		Trigger     WindsurfTriggerType `yaml:"trigger"`
		Globs       string              `yaml:"globs,omitempty"`
		Description string              `yaml:"description,omitempty"`
	}

	WindsurfTriggerType string

	WindsurfPrompt struct {
		Slug    string
		Content string
	}
)

const (
	WindsurfTriggerTypeAlways         WindsurfTriggerType = "always_on"
	WindsurfTriggerTypeGlob           WindsurfTriggerType = "glob"
	WindsurfTriggerTypeAgentRequested WindsurfTriggerType = "model_decision"
	WindsurfTriggerTypeManual         WindsurfTriggerType = "manual"
)

type WindsurfBridge struct{}

func NewWindsurfBridge() domain.AgentBridge[WindsurfRule, WindsurfPrompt] {
	return &WindsurfBridge{}
}

func (bridge *WindsurfBridge) ToAgentRule(rule domain.RuleItem) (WindsurfRule, error) {
	switch rule.Metadata.Attach {
	case domain.AttachTypeAlways:
		return WindsurfRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeAlways,
				Globs:       "",
				Description: "",
			},
		}, nil
	case domain.AttachTypeGlob:
		return WindsurfRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeGlob,
				Globs:       strings.Join(rule.Metadata.Glob, ","),
				Description: "",
			},
		}, nil
	case domain.AttachTypeAgentRequested:
		return WindsurfRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeAgentRequested,
				Globs:       "",
				Description: rule.Metadata.Description,
			},
		}, nil
	case domain.AttachTypeManual:
		return WindsurfRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeManual,
				Globs:       "",
				Description: "",
			},
		}, nil
	default:
		return WindsurfRule{}, fmt.Errorf("unsupported rule attach type: %s", rule.Metadata.Attach)
	}
}

func (bridge *WindsurfBridge) FromAgentRule(rule WindsurfRule) (domain.RuleItem, error) {
	emptyGlobs := make([]string, 0)

	if rule.Metadata.Trigger == WindsurfTriggerTypeAlways {
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

	if rule.Metadata.Trigger == WindsurfTriggerTypeGlob {
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

	if rule.Metadata.Trigger == WindsurfTriggerTypeAgentRequested {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAgentRequested,
				Glob:        emptyGlobs,
				Description: rule.Metadata.Description,
			},
		), nil
	}

	if rule.Metadata.Trigger == WindsurfTriggerTypeManual {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeManual,
				Glob:        emptyGlobs,
				Description: "",
			},
		), nil
	}

	return domain.RuleItem{}, fmt.Errorf("unsupported rule trigger type: %s", rule.Metadata.Trigger)
}

func (bridge *WindsurfBridge) ToAgentPrompt(prompt domain.PromptItem) (WindsurfPrompt, error) {
	return WindsurfPrompt{
		Slug:    prompt.Slug,
		Content: prompt.Content,
	}, nil
}

func (bridge *WindsurfBridge) FromAgentPrompt(prompt WindsurfPrompt) (domain.PromptItem, error) {
	return *domain.NewPromptItem(
		prompt.Slug,
		prompt.Content,
		domain.PromptMetadata{},
	), nil
}

func (rule *WindsurfRule) String() (string, error) {
	metaKeys := 3 // trigger, description, globs
	metaValues := make([]string, 0, metaKeys)

	metaValues = append(metaValues, fmt.Sprintf("trigger: %s", rule.Metadata.Trigger))

	// omit description if empty
	if rule.Metadata.Description != "" {
		description := strings.TrimSpace(rule.Metadata.Description)
		metaValues = append(metaValues, fmt.Sprintf("description: %s", description))
	}

	// omit globs if empty
	if rule.Metadata.Globs != "" {
		globs := strings.TrimSpace(rule.Metadata.Globs)
		metaValues = append(metaValues, fmt.Sprintf("globs: %s", globs))
	}

	frontMatter := fmt.Sprintf("---\n%s\n---\n", strings.Join(metaValues, "\n"))

	// Special case: if the content is empty, we need to return just the front matter
	if rule.Content == "" {
		return frontMatter + "\n", nil
	}

	// Remove trailing newlines from the content, then add one newline at the end
	normalizedContent := strings.TrimRight(rule.Content, "\n")
	result := fmt.Sprintf("%s\n%s", frontMatter, normalizedContent)
	return result + "\n", nil
}

func (prompt *WindsurfPrompt) String() (string, error) {
	return prompt.Content, nil
}
