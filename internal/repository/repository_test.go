package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/ajisai/internal/domain"
	"github.com/sushichan044/ajisai/internal/repository"
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

func TestWritePackage(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	adapter := newMockFileAdapter()
	repo, err := repository.NewPresetRepository(adapter)
	require.NoError(t, err)

	pkg := domain.PresetPackage{
		Name: "test-package",
		Rules: []*domain.RuleItem{
			domain.NewRuleItem("test-rule", "Rule content", domain.RuleMetadata{
				Description: "Test rule",
				Attach:      domain.AttachTypeAlways,
			}),
		},
		Prompts: []*domain.PromptItem{
			domain.NewPromptItem("test-prompt", "Prompt content", domain.PromptMetadata{
				Description: "Test prompt",
			}),
		},
	}

	err = repo.WritePackage("test-namespace", pkg)
	require.NoError(t, err)

	rulePath := filepath.Join(
		tempDir,
		adapter.RulesDir(),
		"test-namespace",
		"test-package",
		"test-rule"+adapter.RuleExtension(),
	)
	promptPath := filepath.Join(
		tempDir,
		adapter.PromptsDir(),
		"test-namespace",
		"test-package",
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
}

func TestClean(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	adapter := newMockFileAdapter()
	repo, err := repository.NewPresetRepository(adapter)
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

	err = repo.Clean(testNamespace)
	require.NoError(t, err)

	_, err = os.Stat(rulesDir)
	require.ErrorIs(t, err, os.ErrNotExist, "Rules directory should be removed")

	_, err = os.Stat(promptsDir)
	assert.ErrorIs(t, err, os.ErrNotExist, "Prompts directory should be removed")
}

func TestReadPackage(t *testing.T) {
	adapter := newMockFileAdapter()
	repo, err := repository.NewPresetRepository(adapter)
	require.NoError(t, err)

	pkg, err := repo.ReadPackage("test-namespace")
	require.NoError(t, err)
	assert.Empty(t, pkg.Name)
	assert.Empty(t, pkg.Rules)
	assert.Empty(t, pkg.Prompts)
}
