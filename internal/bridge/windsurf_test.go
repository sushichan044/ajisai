package bridge_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
)

const longModelDecisionDescription = `Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hell`

func TestWindsurfRule_String(t *testing.T) {
	testCases := []struct {
		name     string
		rule     bridge.WindsurfRule
		expected string
	}{
		{
			name: "Always Rule",
			rule: bridge.WindsurfRule{
				Slug:    "always",
				Content: "Always",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAlways,
					Globs:       "",
					Description: "",
				},
			},
			expected: `---
trigger: always_on
---
Always
`,
		},
		{
			name: "Glob Rule",
			rule: bridge.WindsurfRule{
				Slug:    "go",
				Content: "You should print \"GOGO!\"",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeGlob,
					Globs:       "**/*.go,**/*.yaml",
					Description: "",
				},
			},
			expected: `---
trigger: glob
globs: **/*.go,**/*.yaml
---
You should print "GOGO!"
`,
		},
		{
			name: "Manual Rule",
			rule: bridge.WindsurfRule{
				Slug:    "manual",
				Content: "Content",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeManual,
					Globs:       "",
					Description: "",
				},
			},
			expected: `---
trigger: manual
---
Content
`,
		},
		{
			name: "Model Decision Rule with Long Description",
			rule: bridge.WindsurfRule{
				Slug:    "model-decision",
				Content: "HeyHeyHey",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAgentRequested,
					Globs:       "",
					Description: longModelDecisionDescription,
				},
			},
			expected: `---
trigger: model_decision
description: Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hello!Hell
---
HeyHeyHey
`,
		},
		{
			name: "Content normalization - no trailing newline",
			rule: bridge.WindsurfRule{
				Content: "No trailing newline",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAlways,
					Description: "",
					Globs:       "",
				},
			},
			expected: `---
trigger: always_on
---
No trailing newline
`,
		},
		{
			name: "Content normalization - multiple trailing newlines",
			rule: bridge.WindsurfRule{
				Content: "Multiple trailing newlines\n\n\n",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAlways,
					Description: "",
					Globs:       "",
				},
			},
			expected: `---
trigger: always_on
---
Multiple trailing newlines
`,
		},
		{
			name: "Content normalization - empty content",
			rule: bridge.WindsurfRule{
				Content: "",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAlways,
					Description: "",
					Globs:       "",
				},
			},
			expected: `---
trigger: always_on
---
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bridgeInstance := bridge.NewWindsurfBridge()
			actual, err := bridgeInstance.SerializeAgentRule(tc.rule)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestWindsurfBridge_ToAgentRule(t *testing.T) {
	testCases := []struct {
		name          string
		domainRule    domain.RuleItem
		expectedRule  bridge.WindsurfRule
		expectedError string
	}{
		{
			name: "Always Attach Type",
			domainRule: *domain.NewRuleItem(
				"always",
				"Always",
				domain.RuleMetadata{
					Attach: domain.AttachTypeAlways,
					Globs:  []string{},
				},
			),
			expectedRule: bridge.WindsurfRule{
				Slug:    "always",
				Content: "Always",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAlways,
					Globs:       "",
					Description: "",
				},
			},
		},
		{
			name: "Glob Attach Type",
			domainRule: *domain.NewRuleItem(
				"go",
				"You should print \"GOGO!\"",
				domain.RuleMetadata{
					Attach: domain.AttachTypeGlob,
					Globs:  []string{"**/*.go", "**/*.yaml"},
				},
			),
			expectedRule: bridge.WindsurfRule{
				Slug:    "go",
				Content: "You should print \"GOGO!\"",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeGlob,
					Globs:       "**/*.go,**/*.yaml",
					Description: "",
				},
			},
		},
		{
			name: "Agent Requested Attach Type",
			domainRule: *domain.NewRuleItem(
				"model-decision",
				"HeyHeyHey",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeAgentRequested,
					Globs:       []string{},
					Description: longModelDecisionDescription,
				},
			),
			expectedRule: bridge.WindsurfRule{
				Slug:    "model-decision",
				Content: "HeyHeyHey",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAgentRequested,
					Globs:       "",
					Description: longModelDecisionDescription,
				},
			},
		},
		{
			name: "Manual Attach Type",
			domainRule: *domain.NewRuleItem(
				"manual",
				"Content",
				domain.RuleMetadata{
					Attach: domain.AttachTypeManual,
					Globs:  []string{},
				},
			),
			expectedRule: bridge.WindsurfRule{
				Slug:    "manual",
				Content: "Content",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeManual,
					Globs:       "",
					Description: "",
				},
			},
		},
	}

	bridge := bridge.NewWindsurfBridge()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := bridge.ToAgentRule(tc.domainRule)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRule, actual)
			}
		})
	}
}

func TestWindsurfBridge_FromAgentRule(t *testing.T) {
	testCases := []struct {
		name          string
		agentRule     bridge.WindsurfRule
		expectedRule  domain.RuleItem
		expectedError string
	}{
		{
			name: "Always Trigger Type",
			agentRule: bridge.WindsurfRule{
				Slug:    "always",
				Content: "Always",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger: bridge.WindsurfTriggerTypeAlways,
				},
			},
			expectedRule: *domain.NewRuleItem(
				"always",
				"Always",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeAlways,
					Globs:       []string{},
					Description: "",
				},
			),
		},
		{
			name: "Glob Trigger Type",
			agentRule: bridge.WindsurfRule{
				Slug:    "go",
				Content: "You should print \"GOGO!\"",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger: bridge.WindsurfTriggerTypeGlob,
					Globs:   "**/*.go,**/*.yaml",
				},
			},
			expectedRule: *domain.NewRuleItem(
				"go",
				"You should print \"GOGO!\"",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeGlob,
					Globs:       []string{"**/*.go", "**/*.yaml"},
					Description: "",
				},
			),
		},
		{
			name: "Agent Requested Trigger Type",
			agentRule: bridge.WindsurfRule{
				Slug:    "model-decision",
				Content: "HeyHeyHey",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAgentRequested,
					Description: longModelDecisionDescription,
				},
			},
			expectedRule: *domain.NewRuleItem(
				"model-decision",
				"HeyHeyHey",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeAgentRequested,
					Globs:       []string{},
					Description: longModelDecisionDescription,
				},
			),
		},
		{
			name: "Manual Trigger Type",
			agentRule: bridge.WindsurfRule{
				Slug:    "manual",
				Content: "Content",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger: bridge.WindsurfTriggerTypeManual,
				},
			},
			expectedRule: *domain.NewRuleItem(
				"manual",
				"Content",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeManual,
					Globs:       []string{},
					Description: "",
				},
			),
		},
		{
			name: "Unsupported Trigger Type",
			agentRule: bridge.WindsurfRule{
				Slug: "unsupported",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger: "unsupported_trigger_type",
				},
			},
			expectedError: "unsupported rule trigger type: unsupported_trigger_type",
		},
	}

	bridgeInstance := bridge.NewWindsurfBridge()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := bridgeInstance.FromAgentRule(tc.agentRule)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRule, actual)
			}
		})
	}
}

func TestWindsurfBridge_PromptConversion(t *testing.T) {
	bridgeInstance := bridge.NewWindsurfBridge()

	// Test ToAgentPrompt
	domainPrompt := *domain.NewPromptItem(
		"test-prompt",
		"This is a test prompt.",
		domain.PromptMetadata{},
	)

	actualWindsurfPrompt, err := bridgeInstance.ToAgentPrompt(domainPrompt)
	require.NoError(t, err)
	assert.Equal(t, "test-prompt", actualWindsurfPrompt.Slug)
	assert.Equal(t, "This is a test prompt.", actualWindsurfPrompt.Content)

	// Test FromAgentPrompt
	windsurfPrompt := bridge.WindsurfPrompt{
		Slug:    "windsurf-prompt",
		Content: "This is a Windsurf prompt.",
	}

	expectedDomainPrompt := *domain.NewPromptItem(
		"windsurf-prompt",
		"This is a Windsurf prompt.",
		domain.PromptMetadata{},
	)

	actualDomainPrompt, err := bridgeInstance.FromAgentPrompt(windsurfPrompt)
	require.NoError(t, err)
	assert.Equal(t, expectedDomainPrompt, actualDomainPrompt)
}

func TestWindsurfBridge_SerializeAndDeserializeRule(t *testing.T) {
	testCases := []struct {
		name           string
		rule           bridge.WindsurfRule
		expectedString string
	}{
		{
			name: "Always rule",
			rule: bridge.WindsurfRule{
				Slug:    "always-rule",
				Content: "This is an always-on rule",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAlways,
					Globs:       "",
					Description: "",
				},
			},
			expectedString: `---
trigger: always_on
---
This is an always-on rule
`,
		},
		{
			name: "Glob rule",
			rule: bridge.WindsurfRule{
				Slug:    "glob-rule",
				Content: "This is a glob rule",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeGlob,
					Globs:       "*.go,*.md",
					Description: "",
				},
			},
			expectedString: `---
trigger: glob
globs: *.go,*.md
---
This is a glob rule
`,
		},
		{
			name: "Rule with description",
			rule: bridge.WindsurfRule{
				Slug:    "rule-with-description",
				Content: "This rule has a description",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeAgentRequested,
					Globs:       "",
					Description: "This is a description for the rule",
				},
			},
			expectedString: `---
trigger: model_decision
description: This is a description for the rule
---
This rule has a description
`,
		},
		{
			name: "Empty content rule",
			rule: bridge.WindsurfRule{
				Slug:    "empty-rule",
				Content: "",
				Metadata: bridge.WindsurfRuleMetadata{
					Trigger:     bridge.WindsurfTriggerTypeManual,
					Globs:       "",
					Description: "",
				},
			},
			expectedString: `---
trigger: manual
---
`,
		},
	}

	bridgeInstance := bridge.NewWindsurfBridge()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test serialization
			serialized, err := bridgeInstance.SerializeAgentRule(tc.rule)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedString, serialized)

			// Test deserialization - we compare Content and Metadata independently as Content might have different newlines
			deserialized, err := bridgeInstance.DeserializeAgentRule(tc.rule.Slug, serialized)
			require.NoError(t, err)

			// Normalize content by trimming both leading and trailing whitespace (including newlines)
			deserializedContent := strings.TrimSpace(deserialized.Content)
			expectedContent := strings.TrimSpace(tc.rule.Content)

			assert.Equal(t, tc.rule.Slug, deserialized.Slug)
			assert.Equal(t, expectedContent, deserializedContent)
			assert.Equal(t, tc.rule.Metadata, deserialized.Metadata)
		})
	}
}

func TestWindsurfBridge_DeserializeRuleInvalidFormat(t *testing.T) {
	bridgeInstance := bridge.NewWindsurfBridge()

	// Invalid YAML front matter with completely broken format
	invalidContent := `---
trigger: @invalid-format
---

This is invalid
`

	_, err := bridgeInstance.DeserializeAgentRule("invalid-rule", invalidContent)
	require.Error(t, err)
}

func TestWindsurfBridge_SerializeAndDeserializePrompt(t *testing.T) {
	testCases := []struct {
		name   string
		prompt bridge.WindsurfPrompt
	}{
		{
			name: "Simple prompt",
			prompt: bridge.WindsurfPrompt{
				Slug:    "simple-prompt",
				Content: "This is a simple prompt",
			},
		},
		{
			name: "Prompt with markdown",
			prompt: bridge.WindsurfPrompt{
				Slug:    "markdown-prompt",
				Content: "# Markdown Prompt\n\n- Item 1\n- Item 2\n\n```go\nfunc test() {}\n```",
			},
		},
		{
			name: "Empty prompt",
			prompt: bridge.WindsurfPrompt{
				Slug:    "empty-prompt",
				Content: "",
			},
		},
	}

	bridgeInstance := bridge.NewWindsurfBridge()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test serialization
			serialized, err := bridgeInstance.SerializeAgentPrompt(tc.prompt)
			require.NoError(t, err)
			assert.Equal(t, tc.prompt.Content, serialized)

			// Test deserialization
			deserialized, err := bridgeInstance.DeserializeAgentPrompt(tc.prompt.Slug, serialized)
			require.NoError(t, err)
			assert.Equal(t, tc.prompt, deserialized)
		})
	}
}
