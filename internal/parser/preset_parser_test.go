package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/internal/parser"
)

func TestParsePresetPackage(t *testing.T) {
	testCases := []struct {
		name             string
		presetName       string
		config           *domain.Config
		expectedPackage  *domain.PresetPackage
		expectErr        bool
		useElementsMatch bool
		expectedErrorMsg string
	}{
		{
			name:       "empty_directory",
			presetName: "empty_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"empty_case": {Type: "local", Details: domain.LocalInputSourceDetails{Path: "empty_case"}},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name:    "empty_case",
				Rules:   []*domain.RuleItem{},
				Prompts: []*domain.PromptItem{},
			},
			expectErr: false,
		},
		{
			name:       "parse_single_rule_with_front_matter",
			presetName: "single_rule_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"single_rule_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "single_rule_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name: "single_rule_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"rule1",
						"This is the content of test rule 1.\n",
						domain.RuleMetadata{
							Attach: "manual",
							Globs:  []string{"*.go"},
						},
					),
				},
				Prompts: []*domain.PromptItem{},
			},
			expectErr: false,
		},
		{
			name:       "parse_rule_without_front_matter_from_static_file",
			presetName: "rule_no_frontmatter_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"rule_no_frontmatter_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "rule_no_frontmatter_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name: "rule_no_frontmatter_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"rule_no_frontmatter",
						"This rule has no front matter.\n",
						domain.RuleMetadata{},
					),
				},
				Prompts: []*domain.PromptItem{},
			},
			expectErr: false,
		},
		{
			name:       "rule_with_empty_attach_if_missing_in_frontmatter_from_static_file",
			presetName: "missing_attach_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"missing_attach_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "missing_attach_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name: "missing_attach_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"missing_attach",
						"Content\n",
						domain.RuleMetadata{
							Attach: "", // Missing 'attach' results in empty string
							Globs:  []string{"*.txt"},
						},
					),
				},
				Prompts: []*domain.PromptItem{},
			},
			expectErr: false,
		},
		{
			name:       "rule_with_invalid_front_matter_causes_error_from_static_file",
			presetName: "invalid_fm_rule_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"invalid_fm_rule_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "invalid_fm_rule_case"},
					},
				},
			},
			expectErr:        true,
			expectedErrorMsg: "failed to parse rules",
		},
		{
			name:       "ignore_non_md_files_and_files_outside_target_dirs_from_static_file",
			presetName: "ignore_files_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"ignore_files_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "ignore_files_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name: "ignore_files_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"real_rule",
						"Valid rule content\n",
						domain.RuleMetadata{
							Attach: "always",
						},
					),
				},
				Prompts: []*domain.PromptItem{
					domain.NewPromptItem(
						"empty_prompt",
						"", // Empty file content
						domain.PromptMetadata{},
					),
				},
			},
			expectErr:        false,
			useElementsMatch: true, // Order of items (rule vs prompt) might vary
		},
		{
			name:       "parse_single_prompt_with_front_matter_from_static_file",
			presetName: "single_prompt_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"single_prompt_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "single_prompt_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name: "single_prompt_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"empty_rule",
						"",
						domain.RuleMetadata{},
					),
				},
				Prompts: []*domain.PromptItem{
					domain.NewPromptItem(
						"prompt1",
						"This is the content of test prompt 1.\n",
						domain.PromptMetadata{
							Description: "A sample prompt",
						},
					),
				},
			},
			expectErr:        false,
			useElementsMatch: true,
		},
		{
			name:       "parse_prompt_without_front_matter_from_static_file",
			presetName: "prompt_no_frontmatter_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"prompt_no_frontmatter_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "prompt_no_frontmatter_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name: "prompt_no_frontmatter_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"empty_rule",
						"",
						domain.RuleMetadata{},
					),
				},
				Prompts: []*domain.PromptItem{
					domain.NewPromptItem(
						"prompt_no_frontmatter",
						"This prompt has no front matter.\n",
						domain.PromptMetadata{},
					),
				},
			},
			expectErr:        false,
			useElementsMatch: true,
		},
		{
			name:       "parse_nested_rules_from_static_file",
			presetName: "nested_rules_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"nested_rules_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "nested_rules_case"},
					},
				},
			},
			expectErr: false,
			expectedPackage: &domain.PresetPackage{
				Name: "nested_rules_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"foo/bar", // Slug includes subdirectory
						"This is the content of Bar.\n",
						domain.RuleMetadata{
							Attach: "glob",
							Globs:  []string{"*.go"},
						},
					),
				},
				Prompts: []*domain.PromptItem{},
			},
			useElementsMatch: true,
		},
		{
			name:       "parse_both_rules_and_prompts_with_git_subdir_from_static_file",
			presetName: "both_git_subdir_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"both_git_subdir_case": {
						Type: "git",
						Details: domain.GitInputSourceDetails{
							Repository: "dummy_repo_url", // Not used for path resolution in this test, but good to have
							Directory:  "actual_preset_files",
						},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				Name: "both_git_subdir_case",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"ruleA",
						"Rule A content\n",
						domain.RuleMetadata{
							Attach: "glob",
							Globs:  []string{"*.go"},
						},
					),
				},
				Prompts: []*domain.PromptItem{
					domain.NewPromptItem(
						"promptB",
						"Prompt B content\n",
						domain.PromptMetadata{
							Description: "Prompt B desc",
						},
					),
				},
			},
			expectErr:        false,
			useElementsMatch: true,
		},
		{
			name:       "preset_not_found_in_config",
			presetName: "non_existent_preset",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs:   map[string]domain.InputSource{},
			},
			expectErr:        true,
			expectedErrorMsg: "preset non_existent_preset not found",
		},
		{
			name:             "config_not_in_context",
			presetName:       "some_preset",
			config:           nil, // config is set to nil to simulate it not being in context
			expectErr:        true,
			expectedErrorMsg: "config is nil",
		},
		{
			name:             "invalid_config_type_in_context",
			presetName:       "some_preset",
			config:           nil, // Config is set in the test loop for this specific case
			expectErr:        true,
			expectedErrorMsg: "config is nil",
		},
		{
			name:       "file_read_error_due_to_bad_frontmatter_parse_from_static_file", // Renamed for clarity and static file usage
			presetName: "bad_frontmatter_parse_case",
			config: &domain.Config{
				Settings: domain.Settings{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"bad_frontmatter_parse_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "bad_frontmatter_parse_case"},
					},
				},
			},
			expectErr:        true,
			expectedErrorMsg: "failed to parse rules",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPackage, err := parser.ParsePresetPackage(tc.config, tc.presetName)

			if tc.expectErr {
				require.Error(t, err)
				if tc.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMsg, "Error message mismatch")
				}
			} else {
				assert.NoError(t, err)
				if tc.useElementsMatch {
					assert.Equal(t, tc.expectedPackage.Name, actualPackage.Name, "InputKey mismatch")
					// Compare Rule items
					assert.ElementsMatch(t, tc.expectedPackage.Rules, actualPackage.Rules, "Rule items mismatch")
					// Compare Prompt items
					assert.ElementsMatch(t, tc.expectedPackage.Prompts, actualPackage.Prompts, "Prompt items mismatch")
				} else {
					assert.Equal(t, tc.expectedPackage, actualPackage)
				}
			}
		})
	}
}
