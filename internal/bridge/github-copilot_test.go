package bridge_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/bridge"
	"github.com/sushichan044/ajisai/internal/domain"
)

func TestVSCodeGitHubCopilotBridge_ToAgentRule(t *testing.T) {
	tests := []struct {
		name      string
		ruleItem  domain.RuleItem
		expected  bridge.GitHubCopilotInstruction
		expectErr bool
	}{
		{
			name: "AttachTypeAlways",
			ruleItem: *domain.NewRuleItem(
				"test-always",
				"# Test Always\n\nThis rule is always applied.",
				domain.RuleMetadata{
					Attach: domain.AttachTypeAlways,
					Globs:  []string{},
				},
			),
			expected: bridge.GitHubCopilotInstruction{
				Slug:    "test-always",
				Content: "# Test Always\n\nThis rule is always applied.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: bridge.GitHubCopilotApplyToAllPrimary,
				},
			},
			expectErr: false,
		},
		{
			name: "AttachTypeGlob",
			ruleItem: *domain.NewRuleItem(
				"test-glob",
				"# Test Glob\n\nThis rule applies to specific patterns.",
				domain.RuleMetadata{
					Attach: domain.AttachTypeGlob,
					Globs:  []string{"*.go", "internal/**/*.go"},
				},
			),
			expected: bridge.GitHubCopilotInstruction{
				Slug:    "test-glob",
				Content: "# Test Glob\n\nThis rule applies to specific patterns.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: "*.go,internal/**/*.go",
				},
			},
			expectErr: false,
		},
		{
			name: "AttachTypeManual",
			ruleItem: *domain.NewRuleItem(
				"test-manual",
				"# Test Manual\n\nThis rule is applied manually.",
				domain.RuleMetadata{
					Attach: domain.AttachTypeManual,
					Globs:  []string{},
				},
			),
			expected: bridge.GitHubCopilotInstruction{
				Slug:     "test-manual",
				Content:  "# Test Manual\n\nThis rule is applied manually.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{},
			},
			expectErr: false,
		},
		{
			name: "AttachTypeAgentRequested",
			ruleItem: *domain.NewRuleItem(
				"test-agent-requested",
				"# Test Agent\n\nThis rule is requested by agent.",
				domain.RuleMetadata{
					Attach: domain.AttachTypeAgentRequested,
					Globs:  []string{},
				},
			),
			expected: bridge.GitHubCopilotInstruction{
				Slug:     "test-agent-requested",
				Content:  "# Test Agent\n\nThis rule is requested by agent.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{},
			},
			expectErr: false,
		},
		{
			name: "UnsupportedAttachType",
			ruleItem: *domain.NewRuleItem(
				"test-unsupported",
				"# Test Unsupported\n\nThis rule has unsupported attach type.",
				domain.RuleMetadata{
					Attach: "unsupported",
					Globs:  []string{},
				},
			),
			expected:  bridge.GitHubCopilotInstruction{},
			expectErr: true,
		},
	}

	b := bridge.NewGitHubCopilotBridge()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := b.ToAgentRule(tt.ruleItem)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVSCodeGitHubCopilotBridge_FromAgentRule(t *testing.T) {
	tests := []struct {
		name      string
		rule      bridge.GitHubCopilotInstruction
		expected  domain.RuleItem
		expectErr bool
	}{
		{
			name: "WithApplyToAllPrimary",
			rule: bridge.GitHubCopilotInstruction{
				Slug:    "test-all-primary",
				Content: "# Test All Primary\n\nThis rule applies to all files (primary).",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: bridge.GitHubCopilotApplyToAllPrimary,
				},
			},
			expected: *domain.NewRuleItem(
				"test-all-primary",
				"# Test All Primary\n\nThis rule applies to all files (primary).",
				domain.RuleMetadata{
					Attach: domain.AttachTypeAlways,
					Globs:  []string{},
				},
			),
			expectErr: false,
		},
		{
			name: "WithApplyToAllSecondary",
			rule: bridge.GitHubCopilotInstruction{
				Slug:    "test-all-secondary",
				Content: "# Test All Secondary\n\nThis rule applies to all files (secondary).",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: bridge.GitHubCopilotApplyToAllSecondary,
				},
			},
			expected: *domain.NewRuleItem(
				"test-all-secondary",
				"# Test All Secondary\n\nThis rule applies to all files (secondary).",
				domain.RuleMetadata{
					Attach: domain.AttachTypeAlways,
					Globs:  []string{},
				},
			),
			expectErr: false,
		},
		{
			name: "WithSpecificGlobs",
			rule: bridge.GitHubCopilotInstruction{
				Slug:    "test-specific-globs",
				Content: "# Test Specific Globs\n\nThis rule applies to specific patterns.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: "*.go,internal/**/*.go",
				},
			},
			expected: *domain.NewRuleItem(
				"test-specific-globs",
				"# Test Specific Globs\n\nThis rule applies to specific patterns.",
				domain.RuleMetadata{
					Attach: domain.AttachTypeGlob,
					Globs:  []string{"*.go", "internal/**/*.go"},
				},
			),
			expectErr: false,
		},
		{
			name: "WithEmptyApplyTo",
			rule: bridge.GitHubCopilotInstruction{
				Slug:    "test-empty-apply-to",
				Content: "# Test Empty ApplyTo\n\nThis rule has empty ApplyTo.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: "",
				},
			},
			expected: *domain.NewRuleItem(
				"test-empty-apply-to",
				"# Test Empty ApplyTo\n\nThis rule has empty ApplyTo.",
				domain.RuleMetadata{
					Attach: domain.AttachTypeManual,
					Globs:  []string{},
				},
			),
			expectErr: false,
		},
	}

	b := bridge.NewGitHubCopilotBridge()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := b.FromAgentRule(tt.rule)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVSCodeGitHubCopilotBridge_ToAgentPrompt(t *testing.T) {
	tests := []struct {
		name      string
		prompt    domain.PromptItem
		expected  bridge.GitHubCopilotPrompt
		expectErr bool
	}{
		{
			name: "BasicPrompt",
			prompt: *domain.NewPromptItem(
				"test-prompt",
				"# Test Prompt\n\nThis is a test prompt.",
				domain.PromptMetadata{
					Description: "A test prompt description",
				},
			),
			expected: bridge.GitHubCopilotPrompt{
				Slug:    "test-prompt",
				Content: "# Test Prompt\n\nThis is a test prompt.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Description: "A test prompt description",
					Mode:        bridge.GitHubCopilotInstructionModeAgent,
					Tools:       []string{},
				},
			},
			expectErr: false,
		},
	}

	b := bridge.NewGitHubCopilotBridge()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := b.ToAgentPrompt(tt.prompt)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVSCodeGitHubCopilotBridge_FromAgentPrompt(t *testing.T) {
	tests := []struct {
		name      string
		prompt    bridge.GitHubCopilotPrompt
		expected  domain.PromptItem
		expectErr bool
	}{
		{
			name: "WithDescription",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-prompt",
				Content: "# Test Prompt\n\nThis is a test prompt.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Description: "A test prompt description",
					Mode:        bridge.GitHubCopilotInstructionModeAgent,
					Tools:       []string{"tool1", "tool2"},
				},
			},
			expected: *domain.NewPromptItem(
				"test-prompt",
				"# Test Prompt\n\nThis is a test prompt.",
				domain.PromptMetadata{
					Description: "A test prompt description",
				},
			),
			expectErr: false,
		},
		{
			name: "WithEmptyDescription",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-prompt-empty",
				Content: "# Test Empty\n\nThis prompt has empty description.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Description: "",
					Mode:        bridge.GitHubCopilotInstructionModeAsk,
					Tools:       []string{},
				},
			},
			expected: *domain.NewPromptItem(
				"test-prompt-empty",
				"# Test Empty\n\nThis prompt has empty description.",
				domain.PromptMetadata{
					Description: "",
				},
			),
			expectErr: false,
		},
	}

	b := bridge.NewGitHubCopilotBridge()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := b.FromAgentPrompt(tt.prompt)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGitHubCopilotBridge_SerializeAndDeserializeRule(t *testing.T) {
	testCases := []struct {
		name     string
		rule     bridge.GitHubCopilotInstruction
		expected string
	}{
		{
			name: "Without metadata",
			rule: bridge.GitHubCopilotInstruction{
				Slug:     "test-rule",
				Content:  "This is a test rule without metadata.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{},
			},
			expected: `This is a test rule without metadata.`,
		},
		{
			name: "With ApplyTo",
			rule: bridge.GitHubCopilotInstruction{
				Slug:    "test-with-apply-to",
				Content: "This is a test rule with ApplyTo.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: "*.go,*.md",
				},
			},
			expected: `
---
applyTo: '*.go,*.md'
---
This is a test rule with ApplyTo.
`,
		},
		{
			name: "With ApplyToAll",
			rule: bridge.GitHubCopilotInstruction{
				Slug:    "test-with-apply-to-all",
				Content: "This is a test rule with ApplyToAll.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: bridge.GitHubCopilotApplyToAllPrimary,
				},
			},
			expected: `
---
applyTo: '**'
---
This is a test rule with ApplyToAll.
`,
		},
	}

	bridgeInstance := bridge.NewGitHubCopilotBridge()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test serialization
			serialized, err := bridgeInstance.SerializeAgentRule(tc.rule)
			require.NoError(t, err)

			// Test deserialization - compare fields separately due to newline differences
			deserialized, err := bridgeInstance.DeserializeAgentRule(tc.rule.Slug, serialized)
			require.NoError(t, err)

			// Normalize content by trimming leading and trailing newlines
			deserializedContent := strings.TrimSpace(deserialized.Content)
			expectedContent := strings.TrimSpace(tc.rule.Content)

			assert.Equal(t, tc.rule.Slug, deserialized.Slug)
			assert.Equal(t, expectedContent, deserializedContent)
			assert.Equal(t, tc.rule.Metadata.ApplyTo, deserialized.Metadata.ApplyTo)
		})
	}
}

func TestGitHubCopilotBridge_DeserializeRuleInvalidFormat(t *testing.T) {
	bridgeInstance := bridge.NewGitHubCopilotBridge()

	// Invalid YAML front matter with completely broken format
	invalidContent := `---
applyTo: @invalid-format
---

This is invalid
`

	_, err := bridgeInstance.DeserializeAgentRule("invalid-rule", invalidContent)
	require.Error(t, err)
}

func TestGitHubCopilotBridge_SerializeAndDeserializePrompt(t *testing.T) {
	testCases := []struct {
		name     string
		prompt   bridge.GitHubCopilotPrompt
		expected string
	}{
		{
			name: "Without metadata",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:     "test-prompt",
				Content:  "This is a test prompt without metadata.",
				Metadata: bridge.GitHubCopilotPromptMetadata{},
			},
			expected: `This is a test prompt without metadata.`,
		},
		{
			name: "With description",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-with-description",
				Content: "This is a test prompt with description.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Description: "This is a description.",
				},
			},
			expected: `
---
description: This is a description.
---
This is a test prompt with description.
`,
		},
		{
			name: "With mode",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-with-mode",
				Content: "This is a test prompt with mode.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Mode: bridge.GitHubCopilotInstructionModeAsk,
				},
			},
			expected: `
---
mode: ask
---
This is a test prompt with mode.
`,
		},
		{
			name: "With tools",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-with-tools",
				Content: "This is a test prompt with tools.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Tools: []string{"tool1", "tool2"},
				},
			},
			expected: `
---
tools:
- tool1
- tool2
---
This is a test prompt with tools.
`,
		},
		{
			name: "With all metadata",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-with-all",
				Content: "This is a test prompt with all metadata.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Description: "This is a description.",
					Mode:        bridge.GitHubCopilotInstructionModeEdit,
					Tools:       []string{"tool1", "tool2"},
				},
			},
			expected: `
---
description: This is a description.
mode: edit
tools:
- tool1
- tool2
---
This is a test prompt with all metadata.
`,
		},
	}
	bridgeInstance := bridge.NewGitHubCopilotBridge()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test serialization
			serialized, err := bridgeInstance.SerializeAgentPrompt(tc.prompt)
			require.NoError(t, err)

			// Test deserialization - compare fields separately due to newline differences
			deserialized, err := bridgeInstance.DeserializeAgentPrompt(tc.prompt.Slug, serialized)
			require.NoError(t, err)

			// Normalize content by trimming leading and trailing newlines
			deserializedContent := strings.TrimSpace(deserialized.Content)
			expectedContent := strings.TrimSpace(tc.prompt.Content)

			assert.Equal(t, tc.prompt.Slug, deserialized.Slug)
			assert.Equal(t, expectedContent, deserializedContent)
			assert.Equal(t, tc.prompt.Metadata.Description, deserialized.Metadata.Description)
			assert.Equal(t, tc.prompt.Metadata.Mode, deserialized.Metadata.Mode)
			assert.ElementsMatch(t, tc.prompt.Metadata.Tools, deserialized.Metadata.Tools)
		})
	}
}

func TestGitHubCopilotBridge_DeserializePromptInvalidFormat(t *testing.T) {
	bridgeInstance := bridge.NewGitHubCopilotBridge()

	// Invalid YAML front matter with completely broken format
	invalidContent := `---
mode: @invalid-format
---

This is invalid
`

	_, err := bridgeInstance.DeserializeAgentPrompt("invalid-prompt", invalidContent)
	require.Error(t, err)
}
