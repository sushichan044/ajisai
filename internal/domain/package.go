package domain

import (
	"fmt"
	"path/filepath"
)

type (
	PresetType string

	PresetPackage struct {
		Name   string        // name of the preset package. This value is used as the directory name in the cache.
		Rule   []*RuleItem   // rules in the preset package
		Prompt []*PromptItem // prompts in the preset package
	}

	RuleItem struct {
		presetItem
		Metadata RuleMetadata
	}

	PromptItem struct {
		presetItem
		Metadata PromptMetadata
	}

	// RuleMetadata defines the structure for metadata specific to rules.
	RuleMetadata struct {
		Description string   // Optional: Detailed description from front matter.
		Attach      string   // Required: How the rule is attached ("always", "glob", "manual", "agent-requested" etc.). No default value.
		Glob        []string // Optional: Glob patterns, used when Attach is "glob".
	}

	// PromptMetadata defines the structure for metadata specific to prompts.
	PromptMetadata struct {
		Description string // Optional: Detailed description from front matter.
	}

	// PresetItem is a base struct for all preset items.
	presetItem struct {
		Slug    string // slug of the preset item. (e.g. $preset-name/rules/react/my-rule.md → "react/my-rule")
		Content string // Content (e.g., Markdown), excluding front matter

		Type PresetType // type of the preset item
	}
)

const (
	RulesPresetType   PresetType = "rules"
	PromptsPresetType PresetType = "prompts"

	RuleInternalExtension   = "md"
	PromptInternalExtension = "md"
)

func NewRuleItem(slug string, content string, metadata RuleMetadata) *RuleItem {
	return &RuleItem{
		presetItem: presetItem{
			Type:    RulesPresetType,
			Slug:    slug,
			Content: content,
		},
		Metadata: metadata,
	}
}

func NewPromptItem(slug string, content string, metadata PromptMetadata) *PromptItem {
	return &PromptItem{
		presetItem: presetItem{
			Type:    PromptsPresetType,
			Slug:    slug,
			Content: content,
		},
		Metadata: metadata,
	}
}

func GetItemInternalPath(item presetItem) (string, error) {
	switch item.Type {
	case RulesPresetType:
		return filepath.Join(string(RulesPresetType), item.Slug+"."+RuleInternalExtension), nil
	case PromptsPresetType:
		return filepath.Join(string(PromptsPresetType), item.Slug+"."+PromptInternalExtension), nil
	default:
		return "", fmt.Errorf("unknown preset type: %s", item.Type)
	}
}
