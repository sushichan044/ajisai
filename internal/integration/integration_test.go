package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/internal/integration"
)

type mockFileAdapter struct {
	ruleExt    string
	promptExt  string
	rulesDir   string
	promptsDir string
}

func newMockFileAdapter() *mockFileAdapter {
	return &mockFileAdapter{
		ruleExt:    ".instructions.md",
		promptExt:  ".prompt.md",
		rulesDir:   ".test/instructions",
		promptsDir: ".test/prompts",
	}
}

func (m *mockFileAdapter) RuleExtension() string {
	return m.ruleExt
}

func (m *mockFileAdapter) PromptExtension() string {
	return m.promptExt
}

func (m *mockFileAdapter) RulesDir() string {
	return m.rulesDir
}

func (m *mockFileAdapter) PromptsDir() string {
	return m.promptsDir
}

func (m *mockFileAdapter) SerializeRule(rule *domain.RuleItem) (string, error) {
	return "---\ndescription: " + rule.Metadata.Description + "\nattach: " + string(
		rule.Metadata.Attach,
	) + "\n---\n" + rule.Content, nil
}

func (m *mockFileAdapter) SerializePrompt(prompt *domain.PromptItem) (string, error) {
	return "---\ndescription: " + prompt.Metadata.Description + "\n---\n" + prompt.Content, nil
}

func TestWritePreset(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	adapter := newMockFileAdapter()
	repo, err := integration.New(adapter)
	require.NoError(t, err)

	preset := domain.AgentPreset{
		Name: "test-preset",
		Rules: []*domain.RuleItem{
			domain.NewRuleItem("test-package", "test-preset", "test-rule", "Rule content", domain.RuleMetadata{
				Description: "Test rule",
				Attach:      domain.AttachTypeAlways,
			}),
		},
		Prompts: []*domain.PromptItem{
			domain.NewPromptItem("test-package", "test-preset", "test-prompt", "Prompt content", domain.PromptMetadata{
				Description: "Test prompt",
			}),
		},
	}

	pkg := domain.AgentPresetPackage{
		PackageName: "test-package",
		Presets:     []*domain.AgentPreset{&preset},
	}

	err = repo.WritePackage("test-namespace", &pkg)
	require.NoError(t, err)

	rulePath := filepath.Join(
		tempDir,
		adapter.RulesDir(),
		"test-namespace",
		"test-package",
		"test-preset",
		"test-rule"+adapter.RuleExtension(),
	)
	promptPath := filepath.Join(
		tempDir,
		adapter.PromptsDir(),
		"test-namespace",
		"test-package",
		"test-preset",
		"test-prompt"+adapter.PromptExtension(),
	)

	ruleContent, err := os.ReadFile(rulePath)
	require.NoError(t, err)
	assert.Contains(t, string(ruleContent), "description: Test rule")
	assert.Contains(t, string(ruleContent), "attach: always")
	assert.Contains(t, string(ruleContent), "Rule content")

	promptContent, err := os.ReadFile(promptPath)
	require.NoError(t, err)
	assert.Contains(t, string(promptContent), "description: Test prompt")
	assert.Contains(t, string(promptContent), "Prompt content")

	// Check that .gitignore files are created in namespace directories
	rulesGitignorePath := filepath.Join(tempDir, adapter.RulesDir(), "test-namespace", ".gitignore")
	promptsGitignorePath := filepath.Join(tempDir, adapter.PromptsDir(), "test-namespace", ".gitignore")

	rulesGitignoreContent, err := os.ReadFile(rulesGitignorePath)
	require.NoError(t, err)
	assert.Equal(t, "*\n", string(rulesGitignoreContent), "Rules .gitignore should contain '*'")

	promptsGitignoreContent, err := os.ReadFile(promptsGitignorePath)
	require.NoError(t, err)
	assert.Equal(t, "*\n", string(promptsGitignoreContent), "Prompts .gitignore should contain '*'")
}

func TestClean(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	adapter := newMockFileAdapter()
	repo, err := integration.New(adapter)
	require.NoError(t, err)

	testNamespace := "test-namespace"
	rulesDir := filepath.Join(tempDir, adapter.RulesDir(), testNamespace)
	promptsDir := filepath.Join(tempDir, adapter.PromptsDir(), testNamespace)

	err = os.MkdirAll(rulesDir, 0750)
	require.NoError(t, err)
	err = os.MkdirAll(promptsDir, 0750)
	require.NoError(t, err)

	testRuleFile := filepath.Join(rulesDir, "test-rule.md")
	testPromptFile := filepath.Join(promptsDir, "test-prompt.md")

	err = os.WriteFile(testRuleFile, []byte("test rule content"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(testPromptFile, []byte("test prompt content"), 0600)
	require.NoError(t, err)

	// Create .gitignore files
	testRuleGitignore := filepath.Join(rulesDir, ".gitignore")
	testPromptGitignore := filepath.Join(promptsDir, ".gitignore")
	err = os.WriteFile(testRuleGitignore, []byte("*\n"), 0600)
	require.NoError(t, err)
	err = os.WriteFile(testPromptGitignore, []byte("*\n"), 0600)
	require.NoError(t, err)

	err = repo.Clean(testNamespace)
	require.NoError(t, err)

	_, err = os.Stat(rulesDir)
	require.ErrorIs(t, err, os.ErrNotExist, "Rules directory should be removed")

	_, err = os.Stat(promptsDir)
	assert.ErrorIs(t, err, os.ErrNotExist, "Prompts directory should be removed")
}

func TestEnsureGitignoreFiles(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	adapter := newMockFileAdapter()
	repo, err := integration.New(adapter)
	require.NoError(t, err)

	// Create empty package to trigger gitignore creation
	emptyPreset := domain.AgentPreset{
		Name:    "empty-preset",
		Rules:   []*domain.RuleItem{},
		Prompts: []*domain.PromptItem{},
	}

	pkg := domain.AgentPresetPackage{
		PackageName: "empty-package",
		Presets:     []*domain.AgentPreset{&emptyPreset},
	}

	err = repo.WritePackage("gitignore-test", &pkg)
	require.NoError(t, err)

	// Check that .gitignore files are created with correct content
	rulesGitignorePath := filepath.Join(tempDir, adapter.RulesDir(), "gitignore-test", ".gitignore")
	promptsGitignorePath := filepath.Join(tempDir, adapter.PromptsDir(), "gitignore-test", ".gitignore")

	rulesGitignoreContent, err := os.ReadFile(rulesGitignorePath)
	require.NoError(t, err)
	assert.Equal(t, "*\n", string(rulesGitignoreContent))

	promptsGitignoreContent, err := os.ReadFile(promptsGitignorePath)
	require.NoError(t, err)
	assert.Equal(t, "*\n", string(promptsGitignoreContent))

	// Verify directories exist
	rulesNamespaceDir := filepath.Join(tempDir, adapter.RulesDir(), "gitignore-test")
	promptsNamespaceDir := filepath.Join(tempDir, adapter.PromptsDir(), "gitignore-test")

	_, err = os.Stat(rulesNamespaceDir)
	require.NoError(t, err, "Rules namespace directory should exist")

	_, err = os.Stat(promptsNamespaceDir)
	require.NoError(t, err, "Prompts namespace directory should exist")
}
