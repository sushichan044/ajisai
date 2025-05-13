package utils

import (
	"bytes"

	"github.com/adrg/frontmatter"
)

type (
	ParsedMarkdown[T any] struct {
		// Metadata parsed from the front matter
		FrontMatter T
		// Content is the rest of the markdown content after the front matter
		Content string
	}
)

func ParseMarkdownWithMetadata[T any](content []byte) (ParsedMarkdown[T], error) {
	fm := new(T)
	// Parse the front matter and require it to be present
	rest, err := frontmatter.MustParse(bytes.NewReader(content), fm)
	if err != nil {
		return ParsedMarkdown[T]{}, err
	}

	// Convert the rest of the content to a string
	return ParsedMarkdown[T]{FrontMatter: *fm, Content: string(rest)}, nil
}
