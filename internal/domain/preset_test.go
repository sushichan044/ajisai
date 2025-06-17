package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			result := domain.NewRuleItem("test-package", "test-preset", tt.slug, tt.content, tt.metadata)

			assert.Equal(t, tt.slug, result.URI.Path)
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
			result := domain.NewPromptItem("test-package", "test-preset", tt.slug, tt.content, tt.metadata)

			assert.Equal(t, tt.slug, result.URI.Path)
			assert.Equal(t, tt.content, result.Content)
			assert.Equal(t, domain.PromptsPresetType, result.Type)
			assert.Equal(t, tt.expectedDescription, result.Metadata.Description)
		})
	}
}

func TestAgentPreset_ToXML(t *testing.T) {
	tests := []struct {
		name     string
		preset   *domain.AgentPreset
		expected string
	}{
		{
			name: "Empty preset",
			preset: &domain.AgentPreset{
				Name:    "empty-preset",
				Rules:   nil,
				Prompts: nil,
			},
			expected: `<preset name="empty-preset"></preset>`,
		},
		{
			name: "Preset with rules only",
			preset: &domain.AgentPreset{
				Name: "rules-only",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"test-package",
						"rules-only",
						"rule1",
						"Rule 1 content",
						domain.RuleMetadata{Description: "Rule 1 desc", Attach: domain.AttachTypeAlways},
					),
					domain.NewRuleItem(
						"test-package",
						"rules-only",
						"rule2",
						"Rule 2 content",
						domain.RuleMetadata{
							Description: "Rule 2 desc",
							Attach:      domain.AttachTypeGlob,
							Globs:       []string{"*.go"},
						},
					),
				},
				Prompts: nil,
			},
			expected: `<preset name="rules-only">
  <rules>
    <rule path="rule1">
      <metadata>
        <description>Rule 1 desc</description>
        <attach>always</attach>
      </metadata>
    </rule>
    <rule path="rule2">
      <metadata>
        <description>Rule 2 desc</description>
        <attach>glob</attach>
        <globs>
          <glob>*.go</glob>
        </globs>
      </metadata>
    </rule>
  </rules>
</preset>`,
		},
		{
			name: "Preset with prompts only",
			preset: &domain.AgentPreset{
				Name:  "prompts-only",
				Rules: nil,
				Prompts: []*domain.PromptItem{
					domain.NewPromptItem(
						"test-package",
						"prompts-only",
						"prompt1",
						"Prompt 1 content",
						domain.PromptMetadata{Description: "Prompt 1 desc"},
					),
					domain.NewPromptItem(
						"test-package",
						"prompts-only",
						"prompt2",
						"Prompt 2 content",
						domain.PromptMetadata{Description: "Prompt 2 desc"},
					),
				},
			},
			expected: `<preset name="prompts-only">
  <prompts>
    <prompt path="prompt1">
      <metadata>
        <description>Prompt 1 desc</description>
      </metadata>
    </prompt>
    <prompt path="prompt2">
      <metadata>
        <description>Prompt 2 desc</description>
      </metadata>
    </prompt>
  </prompts>
</preset>`,
		},
		{
			name: "Preset with both rules and prompts",
			preset: &domain.AgentPreset{
				Name: "mixed-preset",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"test-package",
						"mixed-preset",
						"rule1",
						"Rule content",
						domain.RuleMetadata{Description: "Rule desc", Attach: domain.AttachTypeManual},
					),
				},
				Prompts: []*domain.PromptItem{
					domain.NewPromptItem(
						"test-package",
						"mixed-preset",
						"prompt1",
						"Prompt content",
						domain.PromptMetadata{Description: "Prompt desc"},
					),
				},
			},
			expected: `<preset name="mixed-preset">
  <rules>
    <rule path="rule1">
      <metadata>
        <description>Rule desc</description>
        <attach>manual</attach>
      </metadata>
    </rule>
  </rules>
  <prompts>
    <prompt path="prompt1">
      <metadata>
        <description>Prompt desc</description>
      </metadata>
    </prompt>
  </prompts>
</preset>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := tt.preset.MarshalToXML()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(bytes))
		})
	}
}
