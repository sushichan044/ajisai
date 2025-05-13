package bridge_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
)

func TestCursorBridge_RuleConversion(t *testing.T) {
	bridgeInstance := bridge.NewCursorBridge()

	// Test converting domain.RuleItem to CursorRule
	testCases := []struct {
		name         string
		domainRule   domain.RuleItem
		expectedRule bridge.CursorRule
		expectError  bool
	}{
		{
			name: "AlwaysAttach rule",
			domainRule: *domain.NewRuleItem(
				"always-rule",
				"This is an always rule",
				domain.RuleMetadata{
					Attach: domain.AttachTypeAlways,
					Globs:  []string{},
				},
			),
			expectedRule: bridge.CursorRule{
				Slug:    "always-rule",
				Content: "This is an always rule",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: true,
					Description: "",
					Globs:       "",
				},
			},
		},
		{
			name: "Glob rule",
			domainRule: *domain.NewRuleItem(
				"glob-rule",
				"This is a glob rule",
				domain.RuleMetadata{
					Attach: domain.AttachTypeGlob,
					Globs:  []string{"*.go", "*.md"},
				},
			),
			expectedRule: bridge.CursorRule{
				Slug:    "glob-rule",
				Content: "This is a glob rule",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: false,
					Description: "",
					Globs:       "*.go,*.md",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := bridgeInstance.ToAgentRule(tc.domainRule)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRule, actual)
			}
		})
	}

	// Test converting CursorRule to domain.RuleItem
	testCases2 := []struct {
		name         string
		cursorRule   bridge.CursorRule
		expectedRule domain.RuleItem
		expectError  bool
	}{
		{
			name: "AlwaysApply rule",
			cursorRule: bridge.CursorRule{
				Slug:    "always-rule",
				Content: "This is an always rule",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: true,
					Description: "",
					Globs:       "",
				},
			},
			expectedRule: *domain.NewRuleItem(
				"always-rule",
				"This is an always rule",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeAlways,
					Globs:       []string{},
					Description: "",
				},
			),
		},
		{
			name: "Glob rule",
			cursorRule: bridge.CursorRule{
				Slug:    "glob-rule",
				Content: "This is a glob rule",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: false,
					Description: "",
					Globs:       "*.go,*.md",
				},
			},
			expectedRule: *domain.NewRuleItem(
				"glob-rule",
				"This is a glob rule",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeGlob,
					Globs:       []string{"*.go", "*.md"},
					Description: "",
				},
			),
		},
	}

	for _, tc := range testCases2 {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := bridgeInstance.FromAgentRule(tc.cursorRule)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRule, actual)
			}
		})
	}
}

func TestCursorBridge_PromptConversion(t *testing.T) {
	bridgeInstance := bridge.NewCursorBridge()

	// Test ToAgentPrompt
	domainPrompt := *domain.NewPromptItem(
		"test-prompt",
		"This is a test prompt.",
		domain.PromptMetadata{},
	)

	actualCursorPrompt, err := bridgeInstance.ToAgentPrompt(domainPrompt)
	require.NoError(t, err)
	assert.Equal(t, "test-prompt", actualCursorPrompt.Slug)
	assert.Equal(t, "This is a test prompt.", actualCursorPrompt.Content)

	// Test FromAgentPrompt
	cursorPrompt := bridge.CursorPrompt{
		Slug:    "cursor-prompt",
		Content: "This is a Cursor prompt.",
	}

	expectedDomainPrompt := *domain.NewPromptItem(
		"cursor-prompt",
		"This is a Cursor prompt.",
		domain.PromptMetadata{},
	)

	actualDomainPrompt, err := bridgeInstance.FromAgentPrompt(cursorPrompt)
	require.NoError(t, err)
	assert.Equal(t, expectedDomainPrompt, actualDomainPrompt)
}

func TestCursorBridge_SerializeAndDeserializeRule(t *testing.T) {
	testCases := []struct {
		name           string
		rule           bridge.CursorRule
		expectedString string
	}{
		{
			name: "Always apply rule",
			rule: bridge.CursorRule{
				Slug:    "always-rule",
				Content: "This is an always-apply rule",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: true,
					Description: "",
					Globs:       "",
				},
			},
			expectedString: `---
alwaysApply: true
description:
globs:
---
This is an always-apply rule
`,
		},
		{
			name: "Glob rule",
			rule: bridge.CursorRule{
				Slug:    "glob-rule",
				Content: "This is a glob rule",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: false,
					Description: "",
					Globs:       "*.go,*.md",
				},
			},
			expectedString: `---
alwaysApply: false
description:
globs: *.go,*.md
---
This is a glob rule
`,
		},
		{
			name: "Rule with description",
			rule: bridge.CursorRule{
				Slug:    "rule-with-description",
				Content: "This rule has a description",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: false,
					Description: "This is a description for the rule",
					Globs:       "",
				},
			},
			expectedString: `---
alwaysApply: false
description: This is a description for the rule
globs:
---
This rule has a description
`,
		},
		{
			name: "Empty content rule",
			rule: bridge.CursorRule{
				Slug:    "empty-rule",
				Content: "",
				Metadata: bridge.CursorRuleMetadata{
					AlwaysApply: false,
					Description: "",
					Globs:       "",
				},
			},
			expectedString: `---
alwaysApply: false
description:
globs:
---
`,
		},
	}

	bridgeInstance := bridge.NewCursorBridge()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test serialization
			serialized, err := bridgeInstance.SerializeAgentRule(tc.rule)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedString, serialized)

			// Test deserialization - we compare Content and Metadata independently
			deserialized, err := bridgeInstance.DeserializeAgentRule(tc.rule.Slug, serialized)
			require.NoError(t, err)

			// Normalize content by trimming both leading and trailing whitespace
			deserializedContent := strings.TrimSpace(deserialized.Content)
			expectedContent := strings.TrimSpace(tc.rule.Content)

			assert.Equal(t, tc.rule.Slug, deserialized.Slug)
			assert.Equal(t, expectedContent, deserializedContent)
			assert.Equal(t, tc.rule.Metadata, deserialized.Metadata)
		})
	}
}

func TestCursorBridge_DeserializeRuleInvalidFormat(t *testing.T) {
	bridgeInstance := bridge.NewCursorBridge()

	// Invalid YAML front matter with completely broken format
	invalidContent := `---
alwaysApply: @invalid-format
description:
globs:
---

This is invalid
`

	_, err := bridgeInstance.DeserializeAgentRule("invalid-rule", invalidContent)
	require.Error(t, err)
}

func TestCursorBridge_SerializeAndDeserializePrompt(t *testing.T) {
	testCases := []struct {
		name   string
		prompt bridge.CursorPrompt
	}{
		{
			name: "Simple prompt",
			prompt: bridge.CursorPrompt{
				Slug:    "simple-prompt",
				Content: "This is a simple prompt",
			},
		},
		{
			name: "Prompt with markdown",
			prompt: bridge.CursorPrompt{
				Slug:    "markdown-prompt",
				Content: "# Markdown Prompt\n\n- Item 1\n- Item 2\n\n```go\nfunc test() {}\n```",
			},
		},
		{
			name: "Empty prompt",
			prompt: bridge.CursorPrompt{
				Slug:    "empty-prompt",
				Content: "",
			},
		},
	}

	bridgeInstance := bridge.NewCursorBridge()

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
