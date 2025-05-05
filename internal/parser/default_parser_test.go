package parser_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/parser"
)

func TestDefaultParser_ImplementsInterface(t *testing.T) {
	assert.Implements(t, (*domain.ConfigPresetParser)(nil), new(parser.DefaultParser))
}

func TestDefaultParser_Parse(t *testing.T) {
	tests := []struct {
		name           string
		sourceDirSetup func(t *testing.T) string // Returns the path to the source dir
		expectError    bool
		expectPackage  *domain.PresetPackage
		wantErrMsg     string // Optional: check for specific error message substring
	}{
		{
			name: "non-existent source directory",
			sourceDirSetup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			expectError:   true,
			expectPackage: nil,
			wantErrMsg:    "no such file or directory", // Check for underlying os error
		},
		{
			name: "empty source directory",
			sourceDirSetup: func(t *testing.T) string {
				return t.TempDir()
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items:    []domain.PresetItem{},
			},
		},
		{
			name: "parse single rule with front matter",
			sourceDirSetup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "default")
				rulesDir := filepath.Join(dir, "rules")
				assert.NoError(t, os.MkdirAll(rulesDir, 0755))
				content := "---\ntitle: Test Rule 1\nglob:\n  - \"*.go\"\n---\n\nThis is the content of test rule 1.\n"
				assert.NoError(t, os.WriteFile(filepath.Join(rulesDir, "rule1.md"), []byte(content), 0644))
				return dir
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []domain.PresetItem{
					{
						Name:         "rule1",
						Type:         "rule",
						Description:  "This is the content of test rule 1.",
						RelativePath: "rules/rule1.md",
						Metadata: map[string]interface{}{
							"title": "Test Rule 1",
							"glob":  []interface{}{"*.go"},
						},
					},
				},
			},
		},
		{
			name: "parse rule without front matter",
			sourceDirSetup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "default")
				rulesDir := filepath.Join(dir, "rules")
				assert.NoError(t, os.MkdirAll(rulesDir, 0755))
				content := "This rule has no front matter."
				assert.NoError(t, os.WriteFile(filepath.Join(rulesDir, "rule_no_frontmatter.md"), []byte(content), 0644))
				return dir
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []domain.PresetItem{
					{
						Name:         "rule_no_frontmatter",
						Type:         "rule",
						Description:  "This rule has no front matter.",
						RelativePath: "rules/rule_no_frontmatter.md",
						Metadata:     map[string]interface{}{}, // Expect empty map
					},
				},
			},
		},
		{
			name: "skip rule with invalid front matter and log warning",
			sourceDirSetup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "default")
				rulesDir := filepath.Join(dir, "rules")
				assert.NoError(t, os.MkdirAll(rulesDir, 0755))
				content := "---\ntitle: Invalid YAML\ninvalid-yaml: [\n---\n\nContent"
				assert.NoError(t, os.WriteFile(filepath.Join(rulesDir, "invalid_yaml.md"), []byte(content), 0644))
				// Add a valid file to ensure parsing continues
				validContent := "---\ntitle: Valid\n---\nValid content"
				assert.NoError(t, os.WriteFile(filepath.Join(rulesDir, "valid.md"), []byte(validContent), 0644))
				return dir
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []domain.PresetItem{
					{
						Name:         "valid",
						Type:         "rule",
						Description:  "Valid content",
						RelativePath: "rules/valid.md",
						Metadata:     map[string]interface{}{"title": "Valid"},
					},
				}, // invalid_yaml.md should be skipped
			},
			// Note: Log verification happens outside this struct
		},
		{
			name: "ignore non-md files and files outside rules/prompts",
			sourceDirSetup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "default")
				rulesDir := filepath.Join(dir, "rules")
				assert.NoError(t, os.MkdirAll(rulesDir, 0755))
				// Valid rule
				assert.NoError(t, os.WriteFile(filepath.Join(rulesDir, "real_rule.md"), []byte("Valid rule content"), 0644))
				// Ignored files
				assert.NoError(t, os.WriteFile(filepath.Join(rulesDir, "ignoreme.txt"), []byte("text file"), 0644))
				assert.NoError(t, os.WriteFile(filepath.Join(dir, "other.md"), []byte("other md"), 0644))
				return dir
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []domain.PresetItem{
					{
						Name:         "real_rule",
						Type:         "rule",
						Description:  "Valid rule content",
						RelativePath: "rules/real_rule.md",
						Metadata:     map[string]interface{}{},
					},
				},
			},
		},
		{
			name: "parse single prompt with front matter",
			sourceDirSetup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "default")
				promptsDir := filepath.Join(dir, "prompts")
				assert.NoError(t, os.MkdirAll(promptsDir, 0755))
				content := "---\ntitle: Test Prompt 1\ndescription: A sample prompt\n---\n\nThis is the content of test prompt 1.\n"
				assert.NoError(t, os.WriteFile(filepath.Join(promptsDir, "prompt1.md"), []byte(content), 0644))
				return dir
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []domain.PresetItem{
					{
						Name:         "prompt1",
						Type:         "prompt",
						Description:  "This is the content of test prompt 1.", // Content part
						RelativePath: "prompts/prompt1.md",
						Metadata: map[string]interface{}{
							"title":       "Test Prompt 1",
							"description": "A sample prompt",
						},
					},
				},
			},
		},
		{
			name: "parse prompt without front matter",
			sourceDirSetup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "default")
				promptsDir := filepath.Join(dir, "prompts")
				assert.NoError(t, os.MkdirAll(promptsDir, 0755))
				content := "This prompt has no front matter."
				assert.NoError(t, os.WriteFile(filepath.Join(promptsDir, "prompt_no_frontmatter.md"), []byte(content), 0644))
				return dir
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []domain.PresetItem{
					{
						Name:         "prompt_no_frontmatter",
						Type:         "prompt",
						Description:  "This prompt has no front matter.",
						RelativePath: "prompts/prompt_no_frontmatter.md",
						Metadata:     map[string]interface{}{}, // Expect empty map
					},
				},
			},
		},
		{
			name: "parse both rules and prompts",
			sourceDirSetup: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "default")
				rulesDir := filepath.Join(dir, "rules")
				promptsDir := filepath.Join(dir, "prompts")
				assert.NoError(t, os.MkdirAll(rulesDir, 0755))
				assert.NoError(t, os.MkdirAll(promptsDir, 0755))
				// Rule
				ruleContent := "---\ntitle: Rule A\n---\nRule A content"
				assert.NoError(t, os.WriteFile(filepath.Join(rulesDir, "ruleA.md"), []byte(ruleContent), 0644))
				// Prompt
				promptContent := "---\ntitle: Prompt B\n---\nPrompt B content"
				assert.NoError(t, os.WriteFile(filepath.Join(promptsDir, "promptB.md"), []byte(promptContent), 0644))
				return dir
			},
			expectError: false,
			expectPackage: &domain.PresetPackage{
				InputKey: "test-key",
				Items: []domain.PresetItem{
					// Order might vary based on WalkDir, so we check existence later
					{
						Name:         "promptB",
						Type:         "prompt",
						Description:  "Prompt B content",
						RelativePath: "prompts/promptB.md",
						Metadata:     map[string]interface{}{"title": "Prompt B"},
					},
					{
						Name:         "ruleA",
						Type:         "rule",
						Description:  "Rule A content",
						RelativePath: "rules/ruleA.md",
						Metadata:     map[string]interface{}{"title": "Rule A"},
					},
				},
			},
		},
		// More test cases for rules and prompts will be added later
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger to capture output
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelWarn}))
			ctx := context.Background()
			// Inject logger into context (assuming DefaultParser will retrieve it)
			// Note: Actual injection mechanism depends on how DefaultParser gets the logger.
			// For now, we prepare it. The Parse method needs modification later.
			// ctx = context.WithValue(ctx, "logger", logger) // Example placeholder

			sourceDir := tt.sourceDirSetup(t)
			p := parser.NewDefaultParser() // Consider passing logger if needed: parser.NewDefaultParser(logger)
			ctx = context.WithValue(ctx, "logger", logger)

			pkg, err := p.Parse(ctx, "test-key", sourceDir)

			if tt.expectError {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
				// Ensure items slice is not nil for comparison, even if empty
				if tt.expectPackage != nil && tt.expectPackage.Items == nil {
					tt.expectPackage.Items = []domain.PresetItem{}
				}
				if pkg != nil && pkg.Items == nil {
					pkg.Items = []domain.PresetItem{}
				}
				// For the mixed test case, use assert.ElementsMatch since order is not guaranteed
				if tt.name == "parse both rules and prompts" {
					assert.ElementsMatch(t, tt.expectPackage.Items, pkg.Items)
					// Check other fields separately
					assert.Equal(t, tt.expectPackage.InputKey, pkg.InputKey)
				} else {
					assert.Equal(t, tt.expectPackage, pkg)
				}

				// Verify log output for specific tests
				if tt.name == "skip rule with invalid front matter and log warning" {
					assert.Contains(t, logBuf.String(), "WARN")
					assert.Contains(t, logBuf.String(), "Failed to parse front matter")
					assert.Contains(t, logBuf.String(), "invalid_yaml.md")
				}
				// Add checks for other logging scenarios if needed (e.g., read errors)
			}
		})
	}
}
