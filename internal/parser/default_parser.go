package parser

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/mitchellh/mapstructure"
	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

// DefaultParser implements the ConfigPresetParser for the default filesystem structure.
// It looks for 'rules/*.md' and 'prompts/*.md'.
type DefaultParser struct{}

// NewDefaultParser creates a new instance of DefaultParser.
func NewDefaultParser() *DefaultParser {
	return &DefaultParser{}
}

// Parse scans the source directory for rules and prompts and returns a PresetPackage.
// It now accepts inputKey to conform to the ConfigPresetParser interface.
func (p *DefaultParser) Parse(ctx context.Context, inputKey, sourceDir string) (*domain.PresetPackage, error) {
	logger := slog.Default()
	logger.Debug("Parsing source directory", "path", sourceDir, "key", inputKey)

	items := []*domain.PresetItem{} // Initialize as slice of pointers

	_, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source directory does not exist: %s", sourceDir)
		}
		return nil, fmt.Errorf("failed to stat source directory %s: %w", sourceDir, err)
	}

	walkDirFunc := func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			logger.Warn("Error accessing path during walk", "path", path, "error", walkErr)
			return walkErr // Propagate the error to stop walking if critical
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil // Skip directories and non-markdown files
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			logger.Warn("Failed to get relative path", "base", sourceDir, "target", path, "error", err)
			return nil // Skip this file if we can't get a relative path
		}

		var itemType string
		if strings.HasPrefix(relPath, "rules/") {
			itemType = "rule"
			relPath = strings.TrimPrefix(relPath, "rules/")
		} else if strings.HasPrefix(relPath, "prompts/") {
			itemType = "prompt"
			relPath = strings.TrimPrefix(relPath, "prompts/")
		} else {
			// Ignore files not in rules/ or prompts/ subdirectories
			logger.Debug("Ignoring file outside rules/ or prompts/", "path", relPath)
			return nil
		}

		logger.Debug("Processing file", "path", path, "type", itemType)

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			logger.Warn("Failed to read file", "path", path, "error", err)
			return nil // Skip this file
		}

		var fmData map[string]interface{}
		content, err := frontmatter.Parse(bytes.NewReader(contentBytes), &fmData)
		if err != nil {
			logger.Warn("Failed to parse front matter (or read content)", "path", path, "error", err)
			// If front matter fails, treat as if no front matter exists, but log warning.
			// Use the whole file content.
			content = contentBytes
			fmData = make(map[string]interface{}) // Ensure fmData is an empty map
		}

		baseName := filepath.Base(path)
		itemName := strings.TrimSuffix(baseName, ".md")

		item := &domain.PresetItem{ // Create as pointer
			Name:         itemName,
			Description:  strings.TrimSpace(string(content)),
			Type:         itemType,
			RelativePath: relPath,
		}

		// Decode metadata based on item type
		var decodeErr error
		var metadata interface{}

		if itemType == "rule" {
			ruleMeta := domain.RuleMetadata{}
			decodeErr = mapstructure.Decode(fmData, &ruleMeta)
			if decodeErr == nil {
				// Validate required Attach field
				if ruleMeta.Attach == "" {
					logger.Warn("Rule metadata missing required 'attach' field, skipping item", "path", relPath)
					return nil // Skip this item
				}
				// TODO: Add validation for allowed Attach values if needed.
				metadata = ruleMeta
			} else {
				// If decoding itself fails, it might implicitly mean Attach is missing or malformed.
				logger.Warn("Error decoding rule metadata, skipping item", "path", relPath, "error", decodeErr)
				return nil // Skip this item due to decoding error
				// metadata = domain.RuleMetadata{} // Assign empty/default is no longer the strategy
			}
		} else if itemType == "prompt" {
			promptMeta := domain.PromptMetadata{}
			decodeErr = mapstructure.Decode(fmData, &promptMeta)
			if decodeErr != nil {
				slog.Warn("Error decoding prompt metadata", "path", relPath, "error", decodeErr)
				metadata = domain.PromptMetadata{}
			} else {
				metadata = promptMeta
			}
		} else {
			slog.Warn("Unknown item type encountered", "type", itemType, "path", relPath)
			metadata = make(map[string]interface{}) // Fallback
		}

		item.Metadata = metadata
		items = append(items, item) // Append pointer to slice of pointers
		logger.Debug("Parsed item", "name", itemName, "type", itemType, "path", relPath)
		return nil
	}

	err = filepath.WalkDir(sourceDir, walkDirFunc)
	if err != nil {
		logger.Error("Error walking directory", "path", sourceDir, "error", err)
	}

	pkg := &domain.PresetPackage{
		InputKey: inputKey, // Set the inputKey
		Items:    items,    // Assign the collected slice of pointers
	}

	logger.Info("Finished parsing source directory", "path", sourceDir, "item_count", len(pkg.Items))
	return pkg, nil
}
