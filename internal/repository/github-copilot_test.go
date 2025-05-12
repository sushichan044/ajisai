package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/aisync/internal/domain"
	"github.com/sushichan044/aisync/internal/repository"
)

func createTestGitHubCopilotRepository(
	t *testing.T,
	instructionsDir, promptsDir string,
) *repository.GitHubCopilotRepository {
	repo, err := repository.NewGitHubCopilotRepositoryWithPaths(instructionsDir, promptsDir)
	require.NoError(t, err)
	return repo
}

func TestNewGitHubCopilotRepository(t *testing.T) {
	// When creating a new repository
	repo, err := repository.NewGitHubCopilotRepository()

	// Then no error should occur
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestGitHubCopilotRepository_WritePackage(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		pkg       domain.PresetPackage
		setup     func(*testing.T) (string, string)
		validate  func(*testing.T, string, string)
	}{
		{
			name:      "EmptyPackage",
			namespace: "test_namespace",
			pkg: domain.PresetPackage{
				Name:    "empty_package",
				Rules:   []*domain.RuleItem{},
				Prompts: []*domain.PromptItem{},
			},
			setup: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				instructionsDir := filepath.Join(tempDir, "instructions")
				promptsDir := filepath.Join(tempDir, "prompts")
				return instructionsDir, promptsDir
			},
			validate: func(t *testing.T, instructionsDir, promptsDir string) {
				namespaceInstructionsDir := filepath.Join(instructionsDir, "test_namespace")
				namespacePromptsDir := filepath.Join(promptsDir, "test_namespace")

				// Directory structure should not be created for empty packages
				_, err := os.Stat(namespaceInstructionsDir)
				require.ErrorIs(
					t,
					err, os.ErrNotExist,
					"Expected instructions directory not to be created for empty package",
				)

				_, err = os.Stat(namespacePromptsDir)
				assert.ErrorIs(
					t,
					err, os.ErrNotExist,
					"Expected prompts directory not to be created for empty package",
				)
			},
		},
		{
			name:      "WithRulesAndPrompts",
			namespace: "test_namespace",
			pkg: domain.PresetPackage{
				Name: "test_package",
				Rules: []*domain.RuleItem{
					domain.NewRuleItem(
						"test-rule",
						"# Test Rule Content",
						domain.RuleMetadata{
							Attach:      domain.AttachTypeAlways,
							Description: "Test rule description",
						},
					),
				},
				Prompts: []*domain.PromptItem{
					domain.NewPromptItem(
						"test-prompt",
						"# Test Prompt Content",
						domain.PromptMetadata{
							Description: "Test prompt description",
						},
					),
				},
			},
			setup: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				instructionsDir := filepath.Join(tempDir, "instructions")
				promptsDir := filepath.Join(tempDir, "prompts")
				return instructionsDir, promptsDir
			},
			validate: func(t *testing.T, instructionsDir, promptsDir string) {
				// Check if files were created with correct content
				instructionPath := filepath.Join(
					instructionsDir,
					"test_namespace",
					"test_package",
					"test-rule.instructions.md",
				)
				promptPath := filepath.Join(
					promptsDir,
					"test_namespace",
					"test_package",
					"test-prompt.prompt.md",
				)

				// Instruction file should exist and contain frontmatter
				instructionContent, err := os.ReadFile(instructionPath)
				require.NoError(t, err)
				assert.Contains(t, string(instructionContent), "applyTo:")

				// Prompt file should exist
				promptContent, err := os.ReadFile(promptPath)
				require.NoError(t, err)
				assert.Contains(t, string(promptContent), "# Test Prompt Content")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given a test repository
			instructionsDir, promptsDir := tc.setup(t)
			repo := createTestGitHubCopilotRepository(t, instructionsDir, promptsDir)

			// When writing a package
			err := repo.WritePackage(tc.namespace, tc.pkg)

			// Then no error should occur
			require.NoError(t, err)

			// And the expected files and directories should be created
			tc.validate(t, instructionsDir, promptsDir)
		})
	}
}

func TestGitHubCopilotRepository_WritePackage_Error(t *testing.T) {
	// Given a test repository with a read-only directory
	tempDir := t.TempDir()
	instructionsDir := filepath.Join(tempDir, "instructions")
	promptsDir := filepath.Join(tempDir, "prompts")

	// Create directories with no write permissions
	require.NoError(t, os.MkdirAll(instructionsDir, 0555))
	require.NoError(t, os.MkdirAll(promptsDir, 0755)) // We'll keep prompts writable for this test

	repo := createTestGitHubCopilotRepository(t, instructionsDir, promptsDir)

	// When writing a package with a rule
	pkg := domain.PresetPackage{
		Name: "test_package",
		Rules: []*domain.RuleItem{
			domain.NewRuleItem(
				"test-rule",
				"# Test Content",
				domain.RuleMetadata{
					Attach:      domain.AttachTypeAlways,
					Description: "Test rule description",
				},
			),
		},
	}

	// Then an error should occur due to permission issues
	err := repo.WritePackage("test_namespace", pkg)
	assert.Error(t, err, "Expected an error when writing to a read-only directory")
}

func TestGitHubCopilotRepository_ReadPackage(t *testing.T) {
	// Given a test repository
	tempDir := t.TempDir()
	instructionsDir := filepath.Join(tempDir, "instructions")
	promptsDir := filepath.Join(tempDir, "prompts")
	repo := createTestGitHubCopilotRepository(t, instructionsDir, promptsDir)

	// When reading a package
	pkg, err := repo.ReadPackage("test_namespace")

	// Then no error should occur and an empty package should be returned
	require.NoError(t, err)
	assert.Equal(t, domain.PresetPackage{}, pkg, "Expected empty package to be returned")
}

func TestGitHubCopilotRepository_Clean(t *testing.T) {
	// Given a test repository with existing files
	tempDir := t.TempDir()
	instructionsDir := filepath.Join(tempDir, "instructions")
	promptsDir := filepath.Join(tempDir, "prompts")

	namespace := "test_namespace"
	nsInstructionsDir := filepath.Join(instructionsDir, namespace)
	nsPromptsDir := filepath.Join(promptsDir, namespace)

	// Create directories and files
	require.NoError(t, os.MkdirAll(nsInstructionsDir, 0755))
	require.NoError(t, os.MkdirAll(nsPromptsDir, 0755))

	testInstructionFile := filepath.Join(nsInstructionsDir, "test_rule.instructions.md")
	testPromptFile := filepath.Join(nsPromptsDir, "test_prompt.prompt.md")

	require.NoError(t, os.WriteFile(testInstructionFile, []byte("test content"), 0644))
	require.NoError(t, os.WriteFile(testPromptFile, []byte("test content"), 0644))

	repo := createTestGitHubCopilotRepository(t, instructionsDir, promptsDir)

	// When cleaning the namespace
	err := repo.Clean(namespace)

	// Then no error should occur
	require.NoError(t, err)

	// And the directories should be removed
	_, err = os.Stat(nsInstructionsDir)
	require.ErrorIs(t, err, os.ErrNotExist, "Instructions directory should have been removed")

	_, err = os.Stat(nsPromptsDir)
	assert.ErrorIs(t, err, os.ErrNotExist, "Prompts directory should have been removed")
}
