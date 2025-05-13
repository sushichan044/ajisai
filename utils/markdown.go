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
