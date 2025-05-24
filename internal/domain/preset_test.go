package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sushichan044/ajisai/internal/domain"
)

func TestNewRuleItem(t *testing.T) {
	tests := []struct {
		name                string
		slug                string
		content             string
		metadata            domain.RuleMetadata
		expectedDescription string
	}{
		{
			name:    "uses metadata description when provided",
			slug:    "test-rule",
			content: "# Some Heading\nContent here.",
			metadata: domain.RuleMetadata{
				Description: "Explicit description",
				Attach:      domain.AttachTypeAlways,
			},
			expectedDescription: "Explicit description",
		},
		{
			name:    "extracts h1 heading when no description provided",
			slug:    "test-rule",
			content: "# Rule Title\nThis is the rule content.",
			metadata: domain.RuleMetadata{
				Attach: domain.AttachTypeGlob,
				Globs:  []string{"*.go"},
			},
			expectedDescription: "Rule Title",
		},
		{
			name:    "handles empty description and no h1 heading",
			slug:    "test-rule",
			content: "## Some h2 heading\nContent without h1.",
			metadata: domain.RuleMetadata{
				Attach: domain.AttachTypeManual,
			},
			expectedDescription: "",
		},
		{
			name:    "uses first h1 heading when multiple exist",
			slug:    "test-rule",
			content: "# First Heading\nSome content.\n# Second Heading\nMore content.",
			metadata: domain.RuleMetadata{
				Attach: domain.AttachTypeAgentRequested,
			},
			expectedDescription: "First Heading",
		},
		{
			name:    "handles h1 with formatting",
			slug:    "test-rule",
			content: "# Rule with **bold** and *italic*\nContent here.",
			metadata: domain.RuleMetadata{
				Attach: domain.AttachTypeAlways,
			},
			expectedDescription: "Rule with bold and italic",
		},
		{
			name:    "handles frontmatter and h1",
			slug:    "test-rule",
			content: "---\ntitle: Frontmatter Title\n---\n# Content Heading\nContent here.",
			metadata: domain.RuleMetadata{
				Attach: domain.AttachTypeGlob,
				Globs:  []string{"*.ts"},
			},
			expectedDescription: "Content Heading",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.NewRuleItem(tt.slug, tt.content, tt.metadata)

			assert.Equal(t, tt.slug, result.Slug)
			assert.Equal(t, tt.content, result.Content)
			assert.Equal(t, domain.RulesPresetType, result.Type)
			assert.Equal(t, tt.expectedDescription, result.Metadata.Description)

			// Verify other metadata fields are preserved
			assert.Equal(t, tt.metadata.Attach, result.Metadata.Attach)
			assert.Equal(t, tt.metadata.Globs, result.Metadata.Globs)
		})
	}
}

func TestNewPromptItem(t *testing.T) {
	tests := []struct {
		name                string
		slug                string
		content             string
		metadata            domain.PromptMetadata
		expectedDescription string
	}{
		{
			name:    "uses metadata description when provided",
			slug:    "test-prompt",
			content: "# Some Heading\nContent here.",
			metadata: domain.PromptMetadata{
				Description: "Explicit description",
			},
			expectedDescription: "Explicit description",
		},
		{
			name:                "extracts h1 heading when no description provided",
			slug:                "test-prompt",
			content:             "# Prompt Title\nThis is the prompt content.",
			metadata:            domain.PromptMetadata{},
			expectedDescription: "Prompt Title",
		},
		{
			name:                "handles empty description and no h1 heading",
			slug:                "test-prompt",
			content:             "## Some h2 heading\nContent without h1.",
			metadata:            domain.PromptMetadata{},
			expectedDescription: "",
		},
		{
			name:                "uses first h1 heading when multiple exist",
			slug:                "test-prompt",
			content:             "# First Heading\nSome content.\n# Second Heading\nMore content.",
			metadata:            domain.PromptMetadata{},
			expectedDescription: "First Heading",
		},
		{
			name:                "handles h1 with formatting",
			slug:                "test-prompt",
			content:             "# Prompt with **bold** and *italic*\nContent here.",
			metadata:            domain.PromptMetadata{},
			expectedDescription: "Prompt with bold and italic",
		},
		{
			name:                "handles frontmatter and h1",
			slug:                "test-prompt",
			content:             "---\ntitle: Frontmatter Title\n---\n# Content Heading\nContent here.",
			metadata:            domain.PromptMetadata{},
			expectedDescription: "Content Heading",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domain.NewPromptItem(tt.slug, tt.content, tt.metadata)

			assert.Equal(t, tt.slug, result.Slug)
			assert.Equal(t, tt.content, result.Content)
			assert.Equal(t, domain.PromptsPresetType, result.Type)
			assert.Equal(t, tt.expectedDescription, result.Metadata.Description)
		})
	}
}
