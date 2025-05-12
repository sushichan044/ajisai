package bridge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/aisync/internal/bridge"
	"github.com/sushichan044/aisync/internal/domain"
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
			actual, err := tc.rule.String()
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
					Glob:   []string{},
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
					Glob:   []string{"**/*.go", "**/*.yaml"},
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
					Glob:        []string{},
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
					Glob:   []string{},
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
					Glob:        []string{},
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
					Glob:        []string{"**/*.go", "**/*.yaml"},
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
					Glob:        []string{},
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
					Glob:        []string{},
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
