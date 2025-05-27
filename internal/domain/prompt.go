package domain

import (
	"encoding/xml"
	"fmt"

	"github.com/sushichan044/ajisai/utils"
)

type (
	PromptItem struct {
		presetItem
		Metadata PromptMetadata `xml:"metadata"`
	}

	// PromptMetadata defines the structure for metadata specific to prompts.
	PromptMetadata struct {
		Description string `xml:"description,omitempty"`
	}
)

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

func (p *PromptItem) URI(packageName, presetName string) string {
	return fmt.Sprintf("ajisai://%s/%s/%s/%s", packageName, presetName, p.Type, p.Slug)
}

// XML marshalling implementation

type (
	xmlPrompt struct {
		xmlPresetItem
		XMLName  xml.Name          `xml:"prompt"`
		Metadata xmlPromptMetadata `xml:"metadata"`
	}

	xmlPromptMetadata struct {
		Description string `xml:"description,omitempty"`
	}
)

func (p *PromptItem) toXML() *xmlPrompt {
	return &xmlPrompt{
		xmlPresetItem: xmlPresetItem{
			Slug: p.Slug,
		},
		Metadata: xmlPromptMetadata{
			Description: p.Metadata.Description,
		},
	}
}
