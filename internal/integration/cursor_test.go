package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/internal/integration"
)

func TestCursorAdapter_NewCursorAdapter(t *testing.T) {
	// Execute
	adapter := integration.NewCursorAdapter()

	// Verify
	assert.NotNil(t, adapter, "NewCursorAdapter should return non-nil adapter")
}

func TestCursorAdapter_SerializeRule(t *testing.T) {
	// Setup
	adapter := integration.NewCursorAdapter()
	rule := domain.NewRuleItem(
		makeTestURI("test-rule", domain.RulesPresetType),
		"# Test Rule\nThis is a test rule.",
		domain.RuleMetadata{
			Description: "Test Rule Description",
			Attach:      domain.AttachTypeAlways,
			Globs:       []string{"**/*.go"},
		},
	)

	// Execute
	serialized, err := adapter.SerializeRule(rule)

	// Verify
	require.NoError(t, err, "SerializeRule should not return error")
	assert.NotEmpty(t, serialized, "Serialized rule should not be empty")
	assert.Contains(t, serialized, "alwaysApply", "Serialized rule should contain 'alwaysApply' field")
	assert.Contains(t, serialized, "# Test Rule", "Serialized rule should include original content")
}

func TestCursorAdapter_SerializePrompt(t *testing.T) {
	// Setup
	adapter := integration.NewCursorAdapter()
	prompt := domain.NewPromptItem(
		makeTestURI("test-prompt", domain.PromptsPresetType),
		"# Test Prompt\nThis is a test prompt.",
		domain.PromptMetadata{
			Description: "Test Prompt Description",
		},
	)

	// Execute
	serialized, err := adapter.SerializePrompt(prompt)

	// Verify
	require.NoError(t, err, "SerializePrompt should not return error")
	assert.NotEmpty(t, serialized, "Serialized prompt should not be empty")
	assert.Contains(t, serialized, "# Test Prompt", "Serialized prompt should include original content")
}
