package domain_test

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/domain"
)

func TestAgentPresetPackage_MarshalToXML(t *testing.T) {
	tests := []struct {
		name        string
		pkg         *domain.AgentPresetPackage
		expectedXML string
		expectError bool
	}{
		{
			name: "empty package",
			pkg: &domain.AgentPresetPackage{
				PackageName: "test-package",
				Presets:     nil,
			},
			expectedXML: `<package name="test-package"></package>`,
			expectError: false,
		},
		{
			name: "package with empty preset",
			pkg: &domain.AgentPresetPackage{
				PackageName: "test-package",
				Presets: []*domain.AgentPreset{
					{
						Name:    "test-preset",
						Rules:   nil,
						Prompts: nil,
					},
				},
			},
			expectedXML: `<package name="test-package">
  <preset name="test-preset"></preset>
</package>`,
			expectError: false,
		},
		{
			name: "package with preset with rules",
			pkg: &domain.AgentPresetPackage{
				PackageName: "test-package",
				Presets: []*domain.AgentPreset{
					{
						Name: "test-preset",
						Rules: []*domain.RuleItem{
							domain.NewRuleItem(
								domain.URI{
									Scheme:  domain.Scheme,
									Package: "test-package",
									Preset:  "test-preset",
									Type:    domain.RulesPresetType,
									Path:    "test-rule",
								},
								"# Test Rule",
								domain.RuleMetadata{},
							),
						},
						Prompts: nil,
					},
				},
			},
			expectedXML: `<package name="test-package">
  <preset name="test-preset">
    <rules>
      <rule path="test-rule">
        <metadata>
          <description>Test Rule</description>
        </metadata>
      </rule>
    </rules>
  </preset>
</package>`,
			expectError: false,
		},
		{
			name: "package with preset with prompts",
			pkg: &domain.AgentPresetPackage{
				PackageName: "test-package",
				Presets: []*domain.AgentPreset{
					{
						Name:  "test-preset",
						Rules: nil,
						Prompts: []*domain.PromptItem{
							domain.NewPromptItem(
								domain.URI{
									Scheme:  domain.Scheme,
									Package: "test-package",
									Preset:  "test-preset",
									Type:    domain.PromptsPresetType,
									Path:    "test-prompt",
								},
								"# Test Prompt",
								domain.PromptMetadata{},
							),
						},
					},
				},
			},
			expectedXML: `<package name="test-package">
  <preset name="test-preset">
    <prompts>
      <prompt path="test-prompt">
        <metadata>
          <description>Test Prompt</description>
        </metadata>
      </prompt>
    </prompts>
  </preset>
</package>`,
			expectError: false,
		},
		{
			name: "package with multiple presets",
			pkg: &domain.AgentPresetPackage{
				PackageName: "test-package",
				Presets: []*domain.AgentPreset{
					{
						Name: "preset1",
						Rules: []*domain.RuleItem{
							domain.NewRuleItem(
								domain.URI{
									Scheme:  domain.Scheme,
									Package: "test-package",
									Preset:  "preset1",
									Type:    domain.RulesPresetType,
									Path:    "rule1",
								},
								"# Rule One",
								domain.RuleMetadata{},
							),
						},
						Prompts: nil,
					},
					{
						Name:  "preset2",
						Rules: nil,
						Prompts: []*domain.PromptItem{
							domain.NewPromptItem(
								domain.URI{
									Scheme:  domain.Scheme,
									Package: "test-package",
									Preset:  "preset2",
									Type:    domain.PromptsPresetType,
									Path:    "prompt2",
								},
								"# Prompt Two",
								domain.PromptMetadata{},
							),
						},
					},
				},
			},
			expectedXML: `<package name="test-package">
  <preset name="preset1">
    <rules>
      <rule path="rule1">
        <metadata>
          <description>Rule One</description>
        </metadata>
      </rule>
    </rules>
  </preset>
  <preset name="preset2">
    <prompts>
      <prompt path="prompt2">
        <metadata>
          <description>Prompt Two</description>
        </metadata>
      </prompt>
    </prompts>
  </preset>
</package>`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlBytes, err := tt.pkg.MarshalToXML()

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.expectedXML, string(xmlBytes))

			var actualXML, expectedXML interface{}
			require.NoError(t, xml.Unmarshal(xmlBytes, &actualXML))
			require.NoError(t, xml.Unmarshal([]byte(tt.expectedXML), &expectedXML))
			require.Equal(t, expectedXML, actualXML)
		})
	}
}
