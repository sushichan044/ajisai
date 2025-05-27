package domain

import (
	"encoding/xml"
	"fmt"

	"github.com/sushichan044/ajisai/utils"
)

type (
	RuleItem struct {
		presetItem
		Metadata RuleMetadata
	}

	// RuleMetadata defines the structure for metadata specific to rules.
	RuleMetadata struct {
		Description string
		Attach      AttachType
		Globs       []string
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

func (r *RuleItem) URI(packageName, presetName string) string {
	return fmt.Sprintf("ajisai://%s/%s/%s/%s", packageName, presetName, r.Type, r.Slug)
}

// XML marshalling implementation

type (
	xmlRule struct {
		xmlPresetItem
		XMLName  xml.Name        `xml:"rule"`
		Metadata xmlRuleMetadata `xml:"metadata"`
	}

	xmlRuleMetadata struct {
		Attach      AttachType
		Description string
		Globs       []string
	}
)

func (r *RuleItem) toXML() *xmlRule {
	return &xmlRule{
		xmlPresetItem: xmlPresetItem{
			Slug: r.Slug,
		},
		Metadata: xmlRuleMetadata{
			Description: r.Metadata.Description,
			Attach:      r.Metadata.Attach,
			Globs:       r.Metadata.Globs,
		},
	}
}

// Custom XML marshalling for xmlRuleMetadata.
// This is needed because of glob marshalling.
func (m xmlRuleMetadata) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	type GlobsType struct {
		Items []string `xml:"glob"`
	}

	type MetadataInner struct {
		Description string     `xml:"description,omitempty"`
		Attach      AttachType `xml:"attach,omitempty"`
		Globs       *GlobsType `xml:"globs,omitempty"`
	}

	var metadata MetadataInner
	metadata.Description = m.Description
	metadata.Attach = m.Attach

	if len(m.Globs) > 0 {
		metadata.Globs = &GlobsType{
			Items: m.Globs,
		}
	}

	return e.EncodeElement(metadata, xml.StartElement{Name: xml.Name{Local: "metadata"}})
}
