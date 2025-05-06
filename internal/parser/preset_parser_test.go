package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/parser"
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
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"empty_case": {Type: "local", Details: domain.LocalInputSourceDetails{Path: "empty_case"}},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "empty_case",
				Items:    []*domain.PresetItem{},
			},
			expectErr: false,
		},
		{
			name:       "parse_single_rule_with_front_matter",
			presetName: "single_rule_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"single_rule_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "single_rule_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "single_rule_case",
				Items: []*domain.PresetItem{
					{
						Name:    "rule1",
						Type:    domain.RulePresetType,
						Content: "This is the content of test rule 1.\n",
						Metadata: domain.RuleMetadata{
							Title:  "Test Rule 1",
							Attach: "manual",
							Glob:   []string{"*.go"},
						},
						RelativePath: "rules/rule1.md",
					},
				},
			},
			expectErr: false,
		},
		{
			name:       "parse_rule_without_front_matter_from_static_file",
			presetName: "rule_no_frontmatter_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"rule_no_frontmatter_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "rule_no_frontmatter_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "rule_no_frontmatter_case",
				Items: []*domain.PresetItem{
					{
						Name:         "rule_no_frontmatter",
						Type:         domain.RulePresetType,
						Content:      "This rule has no front matter.\n",
						Metadata:     domain.RuleMetadata{},
						RelativePath: "rules/rule_no_frontmatter.md",
					},
				},
			},
			expectErr: false,
		},
		{
			name:       "rule_with_empty_attach_if_missing_in_frontmatter_from_static_file",
			presetName: "missing_attach_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"missing_attach_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "missing_attach_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "missing_attach_case",
				Items: []*domain.PresetItem{
					{
						Name:    "missing_attach",
						Type:    domain.RulePresetType,
						Content: "Content\n",
						Metadata: domain.RuleMetadata{
							Title:  "Missing Attach",
							Attach: "", // Missing 'attach' results in empty string
							Glob:   []string{"*.txt"},
						},
						RelativePath: "rules/missing_attach.md",
					},
				},
			},
			expectErr: false,
		},
		{
			name:       "rule_with_invalid_front_matter_causes_error_from_static_file",
			presetName: "invalid_fm_rule_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
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
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"ignore_files_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "ignore_files_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "ignore_files_case",
				Items: []*domain.PresetItem{
					{
						Name:    "real_rule",
						Type:    domain.RulePresetType,
						Content: "Valid rule content\n",
						Metadata: domain.RuleMetadata{
							Attach: "always",
						},
						RelativePath: "rules/real_rule.md",
					},
					{
						Name:         "empty_prompt",
						Type:         domain.PromptPresetType,
						Content:      "", // Empty file content
						Metadata:     domain.PromptMetadata{},
						RelativePath: "prompts/empty_prompt.md",
					},
				},
			},
			expectErr:        false,
			useElementsMatch: true, // Order of items (rule vs prompt) might vary
		},
		{
			name:       "parse_single_prompt_with_front_matter_from_static_file",
			presetName: "single_prompt_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"single_prompt_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "single_prompt_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "single_prompt_case",
				Items: []*domain.PresetItem{
					{
						Name:    "prompt1",
						Type:    domain.PromptPresetType,
						Content: "This is the content of test prompt 1.\n",
						Metadata: domain.PromptMetadata{
							Description: "A sample prompt",
						},
						RelativePath: "prompts/prompt1.md",
					},
					{
						Name:         "empty_rule",
						Type:         domain.RulePresetType,
						Content:      "",
						Metadata:     domain.RuleMetadata{},
						RelativePath: "rules/empty_rule.md",
					},
				},
			},
			expectErr:        false,
			useElementsMatch: true,
		},
		{
			name:       "parse_prompt_without_front_matter_from_static_file",
			presetName: "prompt_no_frontmatter_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"prompt_no_frontmatter_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "prompt_no_frontmatter_case"},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "prompt_no_frontmatter_case",
				Items: []*domain.PresetItem{
					{
						Name:         "prompt_no_frontmatter",
						Type:         domain.PromptPresetType,
						Content:      "This prompt has no front matter.\n",
						Metadata:     domain.PromptMetadata{},
						RelativePath: "prompts/prompt_no_frontmatter.md",
					},
					{
						Name:         "empty_rule",
						Type:         domain.RulePresetType,
						Content:      "",
						Metadata:     domain.RuleMetadata{},
						RelativePath: "rules/empty_rule.md",
					},
				},
			},
			expectErr:        false,
			useElementsMatch: true,
		},
		{
			name:       "parse_nested_rules_from_static_file",
			presetName: "nested_rules_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"nested_rules_case": {
						Type:    "local",
						Details: domain.LocalInputSourceDetails{Path: "nested_rules_case"},
					},
				},
			},
			expectErr: false,
			expectedPackage: &domain.PresetPackage{
				InputKey: "nested_rules_case",
				Items: []*domain.PresetItem{
					{
						Name:    "bar",
						Type:    domain.RulePresetType,
						Content: "This is the content of Bar.\n",
						Metadata: domain.RuleMetadata{
							Title:  "Bar",
							Attach: "glob",
							Glob:   []string{"*.go"}},
						RelativePath: "rules/foo/bar.md",
					},
				},
			},
			useElementsMatch: true,
		},
		{
			name:       "parse_both_rules_and_prompts_with_git_subdir_from_static_file",
			presetName: "both_git_subdir_case",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{
					"both_git_subdir_case": {
						Type: "git",
						Details: domain.GitInputSourceDetails{
							Repository: "dummy_repo_url", // Not used for path resolution in this test, but good to have
							SubDir:     "actual_preset_files",
						},
					},
				},
			},
			expectedPackage: &domain.PresetPackage{
				InputKey: "both_git_subdir_case",
				Items: []*domain.PresetItem{
					{
						Name:    "promptB",
						Type:    domain.PromptPresetType,
						Content: "Prompt B content\n",
						Metadata: domain.PromptMetadata{
							Description: "Prompt B desc",
						},
						RelativePath: "prompts/promptB.md",
					},
					{
						Name:    "ruleA",
						Type:    domain.RulePresetType,
						Content: "Rule A content\n",
						Metadata: domain.RuleMetadata{
							Title:  "Rule A",
							Attach: "glob",
							Glob:   []string{"*.go"},
						},
						RelativePath: "rules/ruleA.md",
					},
				},
			},
			expectErr:        false,
			useElementsMatch: true,
		},
		{
			name:       "preset_not_found_in_config",
			presetName: "non_existent_preset",
			config: &domain.Config{
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
				Inputs: map[string]domain.InputSource{},
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
				Global: domain.GlobalConfig{CacheDir: "testdata/parse_preset_package"},
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
					assert.Equal(t, tc.expectedPackage.InputKey, actualPackage.InputKey, "InputKey mismatch")
					assert.ElementsMatch(t, tc.expectedPackage.Items, actualPackage.Items, "Items mismatch")
				} else {
					assert.Equal(t, tc.expectedPackage, actualPackage)
				}
			}
		})
	}
}
