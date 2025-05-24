package domain

import (
	"path/filepath"

	"github.com/sushichan044/ajisai/utils"
)

const (
	RulesPresetType   PresetType = "rules"
	PromptsPresetType PresetType = "prompts"

	RuleInternalExtension   = ".md"
	PromptInternalExtension = ".md"

	AttachTypeAlways         AttachType = "always"
	AttachTypeGlob           AttachType = "glob"
	AttachTypeAgentRequested AttachType = "agent-requested"
	AttachTypeManual         AttachType = "manual"
)

type (
	PresetType string
	AttachType string

	AgentPresetPackage struct {
		PackageName string
		Presets     []*AgentPreset
	}

	AgentPreset struct {
		Name    string        // name of the preset. This value is used as the directory name in the cache.
		Rules   []*RuleItem   // rules in the preset
		Prompts []*PromptItem // prompts in the preset
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
		Description string     // Optional: Detailed description from front matter.
		Attach      AttachType // Required: How the rule is attached
		Globs       []string   // Optional: Glob patterns, used when Attach is "glob".
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

func NewRuleItem(slug string, content string, metadata RuleMetadata) *RuleItem {
	var resolvedDescription string
	if metadata.Description != "" {
		resolvedDescription = metadata.Description
	} else {
		// Extract h1 heading from content if description is not provided
		resolvedDescription = utils.ExtractH1Heading(content)
	}

	// Update the metadata with the resolved description
	resolvedMetadata := metadata
	resolvedMetadata.Description = resolvedDescription

	return &RuleItem{
		presetItem: presetItem{
			Type:    RulesPresetType,
			Slug:    slug,
			Content: content,
		},
		Metadata: resolvedMetadata,
	}
}

func NewPromptItem(slug string, content string, metadata PromptMetadata) *PromptItem {
	var resolvedDescription string
	if metadata.Description != "" {
		resolvedDescription = metadata.Description
	} else {
		// Extract h1 heading from content if description is not provided
		resolvedDescription = utils.ExtractH1Heading(content)
	}

	// Update the metadata with the resolved description
	resolvedMetadata := metadata
	resolvedMetadata.Description = resolvedDescription

	return &PromptItem{
		presetItem: presetItem{
			Type:    PromptsPresetType,
			Slug:    slug,
			Content: content,
		},
		Metadata: resolvedMetadata,
	}
}

func (item *presetItem) GetInternalPath(packageName, presetName, extension string) (string, error) {
	return filepath.Join(packageName, presetName, item.Slug+extension), nil
}
