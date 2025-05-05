package parser_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/parser"
)

func TestDefaultParser_ImplementsInterface(t *testing.T) {
	assert.Implements(t, (*domain.ConfigPresetParser)(nil), new(parser.DefaultParser))
}

func TestDefaultParser_Parse(t *testing.T) {
	testCases := []struct {
		name             string
		setup            func(t *testing.T, testDir string)
		expectedPackage  *domain.PresetPackage
		expectErr        bool
		useElementsMatch bool
	}{
		{
			name: "empty_directory",
			setup: func(t *testing.T, testDir string) {
				// Create empty rules/ and prompts/ directories
				require.NoError(t, os.MkdirAll(filepath.Join(testDir, "rules"), 0755))
				require.NoError(t, os.MkdirAll(filepath.Join(testDir, "prompts"), 0755))
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items:    []*domain.PresetItem{}, // Changed to slice of pointers
			},
			expectErr: false,
		},
		{
			name: "parse_single_rule_with_front_matter",
			setup: func(t *testing.T, testDir string) {
				createTestFile(t, testDir, "rules/rule1.md", "---\ntitle: Test Rule 1\nattach: manual\nglob:\n  - \"*.go\"\n---\nThis is the content of test rule 1.")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []*domain.PresetItem{
					{
						Name:         "rule1",
						Type:         "rule",
						Description:  "This is the content of test rule 1.",
						RelativePath: "rule1.md",
						Metadata: domain.RuleMetadata{
							Title:  "Test Rule 1",
							Attach: "manual",
							Glob:   []string{"*.go"},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "parse_rule_without_front_matter",
			setup: func(t *testing.T, testDir string) {
				// This rule file lacks the mandatory 'attach' field.
				createTestFile(t, testDir, "rules/rule_no_frontmatter.md", "This rule has no front matter.")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items:    []*domain.PresetItem{}, // Expect empty items as the rule should be skipped.
			},
			expectErr: false,
		},
		{
			name: "rule_with_missing_attach_field_is_skipped", // New test case
			setup: func(t *testing.T, testDir string) {
				createTestFile(t, testDir, "rules/missing_attach.md", "---\ntitle: Missing Attach\nglob: [\"*.txt\"]\n---\nContent")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items:    []*domain.PresetItem{}, // Expect empty items as the rule should be skipped.
			},
			expectErr: false,
		},
		{
			name: "skip_rule_with_invalid_front_matter_and_log_warning",
			setup: func(t *testing.T, testDir string) {
				// invalid_yaml.md should be skipped due to frontmatter parsing error
				createTestFile(t, testDir, "rules/invalid_yaml.md", "---\ntitle: Invalid\n  bad_indent: true\n---\nContent")
				// valid.md should be parsed correctly as it has 'attach'
				createTestFile(t, testDir, "rules/valid.md", "---\ntitle: Valid\nattach: manual\n---\nValid content")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []*domain.PresetItem{
					// Only the valid item should remain
					{
						Name:         "valid",
						Type:         "rule",
						Description:  "Valid content",
						RelativePath: "valid.md",
						Metadata: domain.RuleMetadata{
							Title:  "Valid",
							Attach: "manual",
						},
					},
				},
			},
			expectErr: false,
			// useElementsMatch: true, // No longer needed as only one item is expected
		},
		{
			name: "ignore_non-md_files_and_files_outside_rules/prompts",
			setup: func(t *testing.T, testDir string) {
				// real_rule.md lacks 'attach', so it will be skipped
				createTestFile(t, testDir, "rules/real_rule.md", "Valid rule content")
				createTestFile(t, testDir, "rules/ignore.txt", "Ignore me")
				require.NoError(t, os.MkdirAll(filepath.Join(testDir, "other_dir"), 0755))
				createTestFile(t, testDir, "other_dir/nested.md", "Ignore me too")
				// Add a valid rule with attach to ensure something is parsed
				createTestFile(t, testDir, "rules/another_rule.md", "---\nattach: always\n---\nAnother rule")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []*domain.PresetItem{
					{
						Name:         "another_rule", // Only the valid rule remains
						Type:         "rule",
						Description:  "Another rule",
						RelativePath: "another_rule.md",
						Metadata: domain.RuleMetadata{
							Attach: "always",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "parse_single_prompt_with_front_matter",
			setup: func(t *testing.T, testDir string) {
				createTestFile(t, testDir, "prompts/prompt1.md", "---\ntitle: Test Prompt 1\ndescription: A sample prompt\n---\nThis is the content of test prompt 1.")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []*domain.PresetItem{
					{
						Name:         "prompt1",
						Type:         "prompt",
						Description:  "This is the content of test prompt 1.",
						RelativePath: "prompt1.md",
						Metadata: domain.PromptMetadata{
							Description: "A sample prompt",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "parse_prompt_without_front_matter",
			setup: func(t *testing.T, testDir string) {
				createTestFile(t, testDir, "prompts/prompt_no_frontmatter.md", "This prompt has no front matter.")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []*domain.PresetItem{
					{
						Name:         "prompt_no_frontmatter",
						Type:         "prompt",
						Description:  "This prompt has no front matter.",
						RelativePath: "prompt_no_frontmatter.md",
						Metadata:     domain.PromptMetadata{},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "parse_both_rules_and_prompts",
			setup: func(t *testing.T, testDir string) {
				// ruleA now includes the required 'attach' field
				createTestFile(t, testDir, "rules/ruleA.md", "---\ntitle: Rule A\nattach: glob\nglob: [\"*.go\"]\n---\nRule A content")
				createTestFile(t, testDir, "prompts/promptB.md", "---\ntitle: Prompt B\n---\nPrompt B content")
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []*domain.PresetItem{
					{
						Name:         "promptB",
						Type:         "prompt",
						Description:  "Prompt B content",
						RelativePath: "promptB.md",
						Metadata:     domain.PromptMetadata{},
					},
					{
						Name:         "ruleA",
						Type:         "rule",
						Description:  "Rule A content",
						RelativePath: "ruleA.md",
						Metadata: domain.RuleMetadata{
							Title:  "Rule A",
							Attach: "glob",
							Glob:   []string{"*.go"},
						},
					},
				},
			},
			expectErr:        false,
			useElementsMatch: true,
		},
		// Add more test cases: missing directories, file read errors, etc.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tc.setup(t, tempDir)

			p := parser.NewDefaultParser()
			actualPackage, err := p.Parse(context.Background(), "test-key", tempDir)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.useElementsMatch {
					assert.Equal(t, tc.expectedPackage.InputKey, actualPackage.InputKey, "InputKey mismatch")
					assert.ElementsMatch(t, tc.expectedPackage.Items, actualPackage.Items, "Items mismatch")
				} else {
					assert.Equal(t, tc.expectedPackage, actualPackage)
				}
			}

			// TODO: Add logging output checks if necessary
		})
	}
}

func createTestFile(t *testing.T, baseDir, relPath, content string) {
	fullPath := filepath.Join(baseDir, relPath)
	dir := filepath.Dir(fullPath)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
}
