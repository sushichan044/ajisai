package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	yaml "github.com/goccy/go-yaml"

	"github.com/sushichan044/ai-rules-manager/internal/domain"
	"github.com/sushichan044/ai-rules-manager/internal/utils"
)

type CursorRule struct {
	Slug     string
	Content  string
	Metadata CursorRuleMetadata
}

type CursorRuleMetadata struct {
	AlwaysApply bool   `yaml:"alwaysApply"`
	Description string `yaml:"description"`
	Globs       string `yaml:"globs"` // e.g. "**/*.{js,ts,jsx,tsx}"
}

type CursorPrompt struct {
	Slug    string
	Content string
}

const (
	CursorRuleExtension   = "mdc"
	CursorPromptExtension = "md"
)

var _ domain.AgentAdapter[CursorRule, CursorPrompt] = &CursorAdapter{}

type CursorAdapter struct{}

func NewCursorAdapter() *CursorAdapter {
	return &CursorAdapter{}
}

func (adapter *CursorAdapter) ToAgentRule(rule domain.RuleItem) (CursorRule, error) {
	switch rule.Metadata.Attach {
	case domain.AttachTypeAlways:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: true,
				Description: "",
				Globs:       "",
			},
		}, nil
	case domain.AttachTypeGlob:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: false,
				Description: "",
				Globs:       strings.Join(rule.Metadata.Glob, ","),
			},
		}, nil
	case domain.AttachTypeAgentRequested:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: false,
				Description: rule.Metadata.Description,
				Globs:       "",
			},
		}, nil
	case domain.AttachTypeManual:
		return CursorRule{
			Slug:    rule.Slug,
			Content: rule.Content,
			Metadata: CursorRuleMetadata{
				AlwaysApply: false,
				Description: "",
				Globs:       "",
			},
		}, nil
	default:
		return CursorRule{}, fmt.Errorf("unsupported rule attach type: %s", rule.Metadata.Attach)
	}
}

func (adapter *CursorAdapter) FromAgentRule(rule CursorRule) (domain.RuleItem, error) {
	emptyGlobs := make([]string, 0)

	if rule.Metadata.AlwaysApply {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAlways,
				Glob:        emptyGlobs,
				Description: "",
			},
		), nil
	}

	if rule.Metadata.Globs != "" {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeGlob,
				Glob:        strings.Split(rule.Metadata.Globs, ","),
				Description: "",
			},
		), nil
	}

	if rule.Metadata.Description != "" {
		return *domain.NewRuleItem(
			rule.Slug,
			rule.Content,
			domain.RuleMetadata{
				Attach:      domain.AttachTypeAgentRequested,
				Description: rule.Metadata.Description,
				Glob:        emptyGlobs,
			},
		), nil
	}

	return *domain.NewRuleItem(
		rule.Slug,
		rule.Content,
		domain.RuleMetadata{
			Attach:      domain.AttachTypeManual,
			Description: "",
			Glob:        emptyGlobs,
		},
	), nil
}

func (adapter *CursorAdapter) ToAgentPrompt(prompt domain.PromptItem) (CursorPrompt, error) {
	return CursorPrompt{
		Slug:    prompt.Slug,
		Content: prompt.Content,
	}, nil
}

func (adapter *CursorAdapter) FromAgentPrompt(prompt CursorPrompt) (domain.PromptItem, error) {
	return *domain.NewPromptItem(
		prompt.Slug,
		prompt.Content,
		domain.PromptMetadata{},
	), nil
}

//gocognit:ignore
func (adapter *CursorAdapter) WritePackage(namespace string, pkg domain.PresetPackage) error {
	cwd, wdErr := os.Getwd()
	if wdErr != nil {
		return wdErr
	}
	cursorRoot := filepath.Join(cwd, ".cursor")

	resolveRulePath := func(rule *domain.RuleItem) (string, error) {
		rulePath, err := rule.GetInternalPath(namespace, pkg.Name, CursorRuleExtension)
		if err != nil {
			return "", err
		}

		return filepath.Join(cursorRoot, rulePath), nil
	}

	resolvePromptPath := func(prompt *domain.PromptItem) (string, error) {
		promptPath, err := prompt.GetInternalPath(namespace, pkg.Name, CursorPromptExtension)
		if err != nil {
			return "", err
		}

		return filepath.Join(cursorRoot, promptPath), nil
	}

	eg := errgroup.Group{}

	for _, rule := range pkg.Rule {
		eg.Go(func() error {
			rulePath, err := resolveRulePath(rule)
			if err != nil {
				return err
			}

			cursorRule, ruleConversionErr := adapter.ToAgentRule(*rule)
			if ruleConversionErr != nil {
				return ruleConversionErr
			}

			cursorRuleStr, ruleStrErr := cursorRule.String()
			if ruleStrErr != nil {
				return ruleStrErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(rulePath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for rule %s: %w", rulePath, dirErr)
			}

			return os.WriteFile(rulePath, []byte(cursorRuleStr), 0600)
		})
	}

	for _, prompt := range pkg.Prompt {
		eg.Go(func() error {
			promptPath, err := resolvePromptPath(prompt)
			if err != nil {
				return err
			}

			cursorPrompt, promptConversionErr := adapter.ToAgentPrompt(*prompt)
			if promptConversionErr != nil {
				return promptConversionErr
			}

			cursorPromptStr, promptStrErr := cursorPrompt.String()
			if promptStrErr != nil {
				return promptStrErr
			}

			if dirErr := utils.EnsureDir(filepath.Dir(promptPath)); dirErr != nil {
				return fmt.Errorf("failed to create directory for prompt %s: %w", promptPath, dirErr)
			}

			return os.WriteFile(promptPath, []byte(cursorPromptStr), 0600)
		})
	}

	return eg.Wait()
}

func (adapter *CursorAdapter) ReadPackage(namespace string, pkg domain.PresetPackage) error {
	return nil
}

func (rule *CursorRule) String() (string, error) {
	frontMatterBytes, err := yaml.MarshalWithOptions(rule.Metadata)
	if err != nil {
		return "", err
	}

	fmStr := string(frontMatterBytes)
	var resultLines []string
	lines := strings.Split(strings.TrimRight(fmStr, "\n"), "\n")

	// Cursor only accepts non-standard YAML formatting, so we need to write special encoding logic.
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Handle empty description: `description: ""` -> `description:`
		if trimmedLine == `description: ""` {
			resultLines = append(resultLines, "description:")
			continue
		}

		// Handle empty globs: `globs: ""` -> `globs:`
		if trimmedLine == `globs: ""` {
			resultLines = append(resultLines, "globs:")
			continue
		}

		// Handle non-empty description: `description: "..."` -> `description: '... '`
		// It should be single-quoted, single-line, with a trailing space inside the quotes.
		if strings.HasPrefix(line, "description: ") && rule.Metadata.Description != "" {
			originalDesc := rule.Metadata.Description
			// Replace newlines in original description with spaces to ensure it's single line
			singleLineDesc := strings.ReplaceAll(originalDesc, "\n", " ")
			// Trim any existing trailing spaces from the original description before adding our specific one
			singleLineDesc = strings.TrimRight(singleLineDesc, " ")
			resultLines = append(resultLines, fmt.Sprintf("description: '%s '", singleLineDesc))
			continue
		}

		// Handle non-empty globs: `globs: "content"` -> `globs: content` (remove quotes)
		if strings.HasPrefix(line, "globs: ") && rule.Metadata.Globs != "" {
			content := strings.TrimPrefix(line, "globs: ")
			content = strings.Trim(content, `"`) // Remove surrounding double quotes from default marshalling
			resultLines = append(resultLines, "globs: "+content)
			continue
		}

		resultLines = append(resultLines, line)
	}

	fmStr = strings.Join(resultLines, "\n") + "\n"

	var normalizedContent string
	if rule.Content == "" {
		normalizedContent = ""
	} else {
		normalizedContent = strings.TrimRight(rule.Content, "\n") + "\n"
	}

	finalStr := "---\n" + fmStr + "---\n" + normalizedContent

	return finalStr, nil
}

func (prompt *CursorPrompt) String() (string, error) {
	return prompt.Content, nil
}
