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

func createTestWindsurfRepository(t *testing.T, rulesDir, promptsDir string) *repository.WindsurfRepository {
	repo, err := repository.NewWindsurfRepositoryWithPaths(rulesDir, promptsDir)
	require.NoError(t, err)
	return repo
}

func TestNewWindsurfRepository(t *testing.T) {
	// When creating a new repository
	repo, err := repository.NewWindsurfRepository()

	// Then no error should occur
	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestWindsurfRepository_WritePackage(t *testing.T) {
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
				rulesDir := filepath.Join(tempDir, "rules")
				promptsDir := filepath.Join(tempDir, "prompts")
				return rulesDir, promptsDir
			},
			validate: func(t *testing.T, rulesDir, promptsDir string) {
				namespaceRulesDir := filepath.Join(rulesDir, "test_namespace")
				namespacePromptsDir := filepath.Join(promptsDir, "test_namespace")

				// Directory structure should not be created for empty packages
				_, err := os.Stat(namespaceRulesDir)
				require.ErrorIs(
					t,
					err, os.ErrNotExist,
					"Expected rules directory not to be created for empty package",
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
				rulesDir := filepath.Join(tempDir, "rules")
				promptsDir := filepath.Join(tempDir, "prompts")
				return rulesDir, promptsDir
			},
			validate: func(t *testing.T, rulesDir, promptsDir string) {
				// Check if the right directories were created
				namespacePkgRulesDir := filepath.Join(rulesDir, "test_namespace", "test_package")
				namespacePkgPromptsDir := filepath.Join(promptsDir, "test_namespace", "test_package")

				// At least one rule and one prompt file should exist
				entries, err := os.ReadDir(namespacePkgRulesDir)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(entries), 1, "Expected at least one rule file")

				entries, err = os.ReadDir(namespacePkgPromptsDir)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(entries), 1, "Expected at least one prompt file")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given a test repository
			rulesDir, promptsDir := tc.setup(t)
			repo := createTestWindsurfRepository(t, rulesDir, promptsDir)

			// When writing a package
			err := repo.WritePackage(tc.namespace, tc.pkg)

			// Then no error should occur
			require.NoError(t, err)

			// And the expected files and directories should be created
			tc.validate(t, rulesDir, promptsDir)
		})
	}
}

func TestWindsurfRepository_WritePackage_Error(t *testing.T) {
	// Given a test repository with a read-only directory
	tempDir := t.TempDir()
	rulesDir := filepath.Join(tempDir, "rules")
	promptsDir := filepath.Join(tempDir, "prompts")

	// Create directories with no write permissions
	require.NoError(t, os.MkdirAll(rulesDir, 0555))
	require.NoError(t, os.MkdirAll(promptsDir, 0755)) // We'll keep prompts writable for this test

	repo := createTestWindsurfRepository(t, rulesDir, promptsDir)

	// When writing a package with a rule
	pkg := domain.PresetPackage{
		Name: "test_package",
		Rules: []*domain.RuleItem{
			{
				Metadata: domain.RuleMetadata{
					Attach:      domain.AttachTypeAlways,
					Description: "Test rule description",
				},
			},
		},
	}

	// Then an error should occur due to permission issues
	err := repo.WritePackage("test_namespace", pkg)
	assert.Error(t, err, "Expected an error when writing to a read-only directory")
}

func TestWindsurfRepository_ReadPackage(t *testing.T) {
	// Given a test repository
	tempDir := t.TempDir()
	rulesDir := filepath.Join(tempDir, "rules")
	promptsDir := filepath.Join(tempDir, "prompts")
	repo := createTestWindsurfRepository(t, rulesDir, promptsDir)

	// When reading a package
	pkg, err := repo.ReadPackage("test_namespace")

	// Then no error should occur and an empty package should be returned
	require.NoError(t, err)
	assert.Equal(t, domain.PresetPackage{}, pkg, "Expected empty package to be returned")
}

func TestWindsurfRepository_Clean(t *testing.T) {
	// Given a test repository with existing files
	tempDir := t.TempDir()
	rulesDir := filepath.Join(tempDir, "rules")
	promptsDir := filepath.Join(tempDir, "prompts")

	namespace := "test_namespace"
	nsRulesDir := filepath.Join(rulesDir, namespace)
	nsPromptsDir := filepath.Join(promptsDir, namespace)

	// Create directories and files
	require.NoError(t, os.MkdirAll(nsRulesDir, 0755))
	require.NoError(t, os.MkdirAll(nsPromptsDir, 0755))

	testRuleFile := filepath.Join(nsRulesDir, "test_rule.md")
	testPromptFile := filepath.Join(nsPromptsDir, "test_prompt.md")

	require.NoError(t, os.WriteFile(testRuleFile, []byte("test content"), 0644))
	require.NoError(t, os.WriteFile(testPromptFile, []byte("test content"), 0644))

	repo := createTestWindsurfRepository(t, rulesDir, promptsDir)

	// When cleaning the namespace
	err := repo.Clean(namespace)

	// Then no error should occur
	require.NoError(t, err)

	// And the directories should be removed
	_, err = os.Stat(nsRulesDir)
	require.ErrorIs(t, err, os.ErrNotExist, "Rules directory should have been removed")

	_, err = os.Stat(nsPromptsDir)
	assert.ErrorIs(t, err, os.ErrNotExist, "Prompts directory should have been removed")
}

func TestWindsurfRepository_Clean_NonExistentDir(t *testing.T) {
	// Given a test repository with non-existent directories
	tempDir := t.TempDir()
	rulesDir := filepath.Join(tempDir, "rules")
	promptsDir := filepath.Join(tempDir, "prompts")

	// Don't create the directories, we want to test cleaning non-existent paths

	repo := createTestWindsurfRepository(t, rulesDir, promptsDir)

	// When cleaning the namespace that doesn't exist
	err := repo.Clean("non_existent_namespace")

	// Then no error should occur for attempting to remove non-existent directories
	assert.NoError(t, err)
}
