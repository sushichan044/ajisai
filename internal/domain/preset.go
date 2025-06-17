package domain

import (
	"encoding/xml"
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

	AgentPreset struct {
		Name    string        // name of the preset. This value is used as the directory name in the cache.
		Rules   []*RuleItem   // rules in the preset
		Prompts []*PromptItem // prompts in the preset
	}

	// PresetItem is a base struct for all preset items.
	presetItem struct {
		Content string     `xml:"-"` // Do not output content to XML.
		Type    PresetType `xml:"type,attr"`
		URI     URI        `xml:"-"` // Do not output URI to XML.
	}
)

// XML marshalling implementation

type (
	xmlPreset struct {
		XMLName xml.Name    `xml:"preset"`
		Name    string      `xml:"name,attr"`
		Rules   *xmlRules   `xml:"rules,omitempty"`
		Prompts *xmlPrompts `xml:"prompts,omitempty"`
	}

	xmlPresetItem struct {
		Path string `xml:"path,attr"`
	}

	xmlRules struct {
		Items []*xmlRule `xml:"rule"`
	}

	xmlPrompts struct {
		Items []*xmlPrompt `xml:"prompt"`
	}
)

// MarshalToXML converts the AgentPreset object into its XML representation.
// It returns the XML as a byte slice and an error if the marshalling fails.
func (p *AgentPreset) MarshalToXML() ([]byte, error) {
	return p.toXML().Marshal()
}

func (p *AgentPreset) toXML() *xmlPreset {
	outputPreset := xmlPreset{
		Name: p.Name,
	}

	if len(p.Rules) > 0 {
		items := make([]*xmlRule, len(p.Rules))
		for i, rule := range p.Rules {
			items[i] = rule.toXML()
		}

		outputPreset.Rules = &xmlRules{
			Items: items,
		}
	}

	if len(p.Prompts) > 0 {
		items := make([]*xmlPrompt, len(p.Prompts))
		for i, prompt := range p.Prompts {
			items[i] = prompt.toXML()
		}

		outputPreset.Prompts = &xmlPrompts{
			Items: items,
		}
	}

	return &outputPreset
}

func (xp *xmlPreset) Marshal() ([]byte, error) {
	return xml.MarshalIndent(xp, "", "  ")
}
