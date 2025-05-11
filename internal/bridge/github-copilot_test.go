package bridge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ai-rules-manager/internal/bridge"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
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
					Glob:   []string{},
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
					Glob:   []string{"*.go", "internal/**/*.go"},
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
					Glob:   []string{},
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
					Glob:   []string{},
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
					Glob:   []string{},
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
					Glob:   []string{},
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
					Glob:   []string{},
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
					Glob:   []string{"*.go", "internal/**/*.go"},
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
					Glob:   []string{},
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

func TestVSCodeGitHubCopilotPrompt_String(t *testing.T) {
	tests := []struct {
		name      string
		prompt    bridge.GitHubCopilotPrompt
		expected  string
		expectErr bool
	}{
		{
			name: "WithAllFields",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-prompt",
				Content: "# Test Prompt\n\nThis is a test prompt.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Description: "A test prompt description",
					Mode:        bridge.GitHubCopilotInstructionModeAgent,
					Tools:       []string{"tool1", "tool2"},
				},
			},
			expected: `---
description: A test prompt description
mode: agent
tools:
- tool1
- tool2
---

# Test Prompt

This is a test prompt.`,
			expectErr: false,
		},
		{
			name: "WithEmptyFields",
			prompt: bridge.GitHubCopilotPrompt{
				Slug:    "test-prompt-empty",
				Content: "# Test Empty\n\nThis prompt has empty fields.",
				Metadata: bridge.GitHubCopilotPromptMetadata{
					Description: "",
					Mode:        "",
					Tools:       []string{},
				},
			},
			expected: `# Test Empty

This prompt has empty fields.`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.prompt.String()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVSCodeGitHubCopilotInstruction_String(t *testing.T) {
	tests := []struct {
		name        string
		instruction bridge.GitHubCopilotInstruction
		expected    string
		expectErr   bool
	}{
		{
			name: "WithApplyTo",
			instruction: bridge.GitHubCopilotInstruction{
				Slug:    "test-instruction",
				Content: "# Test Instruction\n\nThis is a test instruction.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: "*.go",
				},
			},
			expected: `---
applyTo: "*.go"
---

# Test Instruction

This is a test instruction.`,
			expectErr: false,
		},
		{
			name: "WithEmptyApplyTo",
			instruction: bridge.GitHubCopilotInstruction{
				Slug:    "test-instruction-empty",
				Content: "# Test Empty\n\nThis instruction has empty ApplyTo.",
				Metadata: bridge.GitHubCopilotInstructionMetadata{
					ApplyTo: "",
				},
			},
			expected: `# Test Empty

This instruction has empty ApplyTo.`,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.instruction.String()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
