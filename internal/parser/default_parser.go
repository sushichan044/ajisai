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
	"github.com/sushichan044/ai-rules-manager/internal/domain"
)

// DefaultParser implements the ConfigPresetParser for the default filesystem structure.
// It looks for 'rules/*.md' and 'prompts/*.md'.
type DefaultParser struct{}

// NewDefaultParser creates a new instance of DefaultParser.
func NewDefaultParser() *DefaultParser {
	return &DefaultParser{}
}

// Parse walks the sourceDir and parses rules and prompts according to the default structure.
func (p *DefaultParser) Parse(ctx context.Context, inputKey, sourceDir string) (*domain.PresetPackage, error) {
	logger := slog.Default() // Default logger
	if ctxLogger, ok := ctx.Value("logger").(*slog.Logger); ok {
		logger = ctxLogger
	}

	items := []domain.PresetItem{}

	_, err := os.Stat(sourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Return specific error if directory doesn't exist
			return nil, fmt.Errorf("source directory '%s' not found: %w", sourceDir, err)
		}
		return nil, fmt.Errorf("failed to stat source directory '%s': %w", sourceDir, err)
	}

	err = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Error accessing path, return it to stop walking if critical, or log and continue
			// For now, let's return the error.
			return fmt.Errorf("error accessing path %q: %w", path, err)
		}

		// Skip the root directory itself
		if path == sourceDir {
			return nil
		}

		// Skip directories
		if d.IsDir() {
			// If we encounter rules/ or prompts/ directly, continue into them.
			// Otherwise, skip other directories entirely.
			relPath, _ := filepath.Rel(sourceDir, path)
			if relPath != "rules" && relPath != "prompts" {
				logger.Debug("Skipping directory", "path", relPath)
				return fs.SkipDir // Skip this directory and its contents
			}
			return nil // Continue into rules/ or prompts/
		}

		// Only process *.md files
		if !strings.HasSuffix(d.Name(), ".md") {
			logger.Debug("Skipping non-markdown file", "path", path)
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			// Should not happen if path is within sourceDir, but handle defensively
			logger.Error("Failed to get relative path", "error", err, "path", path, "sourceDir", sourceDir)
			return fmt.Errorf("failed to get relative path for %s: %w", path, err) // Stop walking
		}

		var itemType string
		if strings.HasPrefix(relPath, "rules/") {
			itemType = "rule"
		} else if strings.HasPrefix(relPath, "prompts/") {
			itemType = "prompt"
		} else {
			logger.Debug("Skipping file outside rules/ or prompts/", "path", relPath)
			return nil // Skip files not in rules/ or prompts/
		}

		// Read file content
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			logger.Error("Failed to read file", "error", err, "path", path)
			// Decide whether to skip or fail. Let's skip and log.
			return nil // Continue walking
		}

		// Parse front matter
		var metadata map[string]interface{}
		contentBody, err := frontmatter.Parse(bytes.NewReader(contentBytes), &metadata)

		// Ensure metadata is non-nil even if parsing fails or no front matter exists
		if metadata == nil {
			metadata = make(map[string]interface{})
		}

		if err != nil {
			logger.Warn("Failed to parse front matter, skipping file", "error", err, "path", relPath)
			return nil // Skip this file entirely if front matter parsing fails
		}

		// Extract name from filename (without extension)
		baseName := filepath.Base(path)
		itemName := strings.TrimSuffix(baseName, ".md")

		item := domain.PresetItem{
			Name:         itemName,
			Type:         itemType,
			Description:  strings.TrimSpace(string(contentBody)), // Use content part as description
			RelativePath: relPath,
			Metadata:     metadata,
		}

		items = append(items, item)
		logger.Debug("Parsed item", "name", itemName, "type", itemType, "path", relPath)

		return nil // Continue walking
	})

	if err != nil {
		// Error during walk
		return nil, fmt.Errorf("error walking directory '%s': %w", sourceDir, err)
	}

	// Return the package with collected items
	pkg := &domain.PresetPackage{
		InputKey: inputKey,
		Items:    items,
	}

	return pkg, nil
}
