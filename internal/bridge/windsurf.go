package bridge

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/utils"
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
			Slug:    rule.URI.Path,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeAlways,
				Globs:       "",
				Description: "",
			},
		}, nil
	case domain.AttachTypeGlob:
		return WindsurfRule{
			Slug:    rule.URI.Path,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeGlob,
				Globs:       strings.Join(rule.Metadata.Globs, ","),
				Description: "",
			},
		}, nil
	case domain.AttachTypeAgentRequested:
		return WindsurfRule{
			Slug:    rule.URI.Path,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeAgentRequested,
				Globs:       "",
				Description: rule.Metadata.Description,
			},
		}, nil
	case domain.AttachTypeManual:
		return WindsurfRule{
			Slug:    rule.URI.Path,
			Content: rule.Content,
			Metadata: WindsurfRuleMetadata{
				Trigger:     WindsurfTriggerTypeManual,
				Globs:       "",
				Description: "",
			},
		}, nil
	}

	// Fallback as manual rule.
	return WindsurfRule{
		Slug:    rule.URI.Path,
		Content: rule.Content,
		Metadata: WindsurfRuleMetadata{
			Trigger:     WindsurfTriggerTypeManual,
			Globs:       "",
			Description: "",
		},
	}, nil
}

func (bridge *WindsurfBridge) FromAgentRule(rule WindsurfRule) (domain.RuleItem, error) {
	emptyGlobs := make([]string, 0)

	// Create URI with placeholder values since bridge doesn't have package/preset context
	uri := domain.URI{
		Scheme:  domain.Scheme,
		Package: "", // placeholder, bridge doesn't have this context
		Preset:  "", // placeholder, bridge doesn't have this context
		Type:    domain.RulesPresetType,
		Path:    rule.Slug,
	}

	if rule.Metadata.Trigger == WindsurfTriggerTypeAlways {
		return *domain.NewRuleItem(
			uri,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAlways,
				Globs:       emptyGlobs,
				Description: "",
			},
		), nil
	}

	if rule.Metadata.Trigger == WindsurfTriggerTypeGlob {
		return *domain.NewRuleItem(
			uri,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeGlob,
				Globs:       strings.Split(rule.Metadata.Globs, ","),
				Description: "",
			},
		), nil
	}

	if rule.Metadata.Trigger == WindsurfTriggerTypeAgentRequested {
		return *domain.NewRuleItem(
			uri,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAgentRequested,
				Globs:       emptyGlobs,
				Description: rule.Metadata.Description,
			},
		), nil
	}

	if rule.Metadata.Trigger == WindsurfTriggerTypeManual {
		return *domain.NewRuleItem(
			uri,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeManual,
				Globs:       emptyGlobs,
				Description: "",
			},
		), nil
	}

	return domain.RuleItem{}, fmt.Errorf("unsupported rule trigger type: %s", rule.Metadata.Trigger)
}

func (bridge *WindsurfBridge) ToAgentPrompt(prompt domain.PromptItem) (WindsurfPrompt, error) {
	return WindsurfPrompt{
		Slug:    prompt.URI.Path,
		Content: prompt.Content,
	}, nil
}

func (bridge *WindsurfBridge) FromAgentPrompt(prompt WindsurfPrompt) (domain.PromptItem, error) {
	// Create URI with placeholder values since bridge doesn't have package/preset context
	uri := domain.URI{
		Scheme:  domain.Scheme,
		Package: "", // placeholder, bridge doesn't have this context
		Preset:  "", // placeholder, bridge doesn't have this context
		Type:    domain.PromptsPresetType,
		Path:    prompt.Slug,
	}

	return *domain.NewPromptItem(
		uri,
		prompt.Content,
		domain.PromptMetadata{},
	), nil
}

func (bridge *WindsurfBridge) SerializeAgentRule(rule WindsurfRule) (string, error) {
	// Windsurf does not accept quoted front matters, so we need to write custom marshaler
	metaKeys := 3 // trigger, description, globs
	metaContent := make([]string, 0, metaKeys)

	metaContent = append(metaContent, "trigger: "+string(rule.Metadata.Trigger))

	// omit description property if empty
	if desc := strings.TrimSpace(rule.Metadata.Description); desc != "" {
		metaContent = append(metaContent, "description: "+desc)
	}

	// omit globs property if empty
	if globs := strings.TrimSpace(rule.Metadata.Globs); globs != "" {
		metaContent = append(metaContent, "globs: "+globs)
	}

	return strings.TrimRight("---\n"+strings.Join(metaContent, "\n")+"\n---\n"+rule.Content, "\n") + "\n", nil
}

func (bridge *WindsurfBridge) DeserializeAgentRule(slug string, ruleBody string) (WindsurfRule, error) {
	lines := strings.Split(ruleBody, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "globs: ") {
			// we need to add quotes around the glob patterns to avoid parsing errors
			lines[i] = "globs: " + strconv.Quote(strings.TrimSpace(line[7:]))
		}
	}
	ruleBody = strings.Join(lines, "\n")

	result, err := utils.ParseMarkdownWithMetadata[WindsurfRuleMetadata]([]byte(ruleBody))
	if err != nil {
		return WindsurfRule{}, err
	}

	return WindsurfRule{
		Slug:     slug,
		Content:  result.Content,
		Metadata: result.FrontMatter,
	}, nil
}

func (bridge *WindsurfBridge) SerializeAgentPrompt(prompt WindsurfPrompt) (string, error) {
	return prompt.Content, nil
}

func (bridge *WindsurfBridge) DeserializeAgentPrompt(slug string, promptBody string) (WindsurfPrompt, error) {
	return WindsurfPrompt{
		Slug:    slug,
		Content: promptBody,
	}, nil
}
