package utils

import (
	"bytes"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type (
	ParsedMarkdown[T any] struct {
		// Metadata parsed from the front matter
		FrontMatter T
		// Content is the rest of the markdown content after the front matter
		Content string
	}
)

// ParseMarkdownWithMetadata parses the front matter from the given markdown content
// and returns the parsed metadata and the rest of the content.
//
// The front matter is expected to be in YAML format and is unmarshalled into the
// provided type T. The rest of the content is returned as a string.
//
// If the front matter is not present, FrontMatter will be empty.
func ParseMarkdownWithMetadata[T any](content []byte) (ParsedMarkdown[T], error) {
	fm := new(T)
	// Parse the front matter and require it to be present
	rest, err := frontmatter.Parse(bytes.NewReader(content), fm)
	if err != nil {
		return ParsedMarkdown[T]{}, err
	}

	// Convert the rest of the content to a string
	return ParsedMarkdown[T]{FrontMatter: *fm, Content: string(rest)}, nil
}

// ExtractH1Heading extracts the first h1 heading from the given markdown content.
// Returns the text content of the h1 heading, or empty string if no h1 heading is found.
func ExtractH1Heading(content string) string {
	md := goldmark.New()
	source := []byte(content)
	reader := text.NewReader(source)

	document := md.Parser().Parse(reader)

	var h1Text string
	walkErr := ast.Walk(document, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
			h1Text = extractTextFromNode(heading, source)
			return ast.WalkStop, nil
		}

		return ast.WalkContinue, nil
	})

	// Gracefully degrade
	if walkErr != nil {
		return ""
	}

	return h1Text
}

// extractTextFromNode recursively extracts all text content from an AST node.
func extractTextFromNode(node ast.Node, source []byte) string {
	var buf strings.Builder

	walkErr := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if textNode, ok := n.(*ast.Text); ok {
			buf.Write(textNode.Segment.Value(source))
		}

		return ast.WalkContinue, nil
	})

	// Gracefully degrade
	if walkErr != nil {
		return ""
	}

	return buf.String()
}
