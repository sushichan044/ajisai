---
attach: glob
globs:
  - "**/*.go"
---

# Local Rules for Go

- Use `errors.Is` for error comparison.
- In test runner, use `require.NoError` for ensure operation success.
- Use `utils.AtomicWriteFile` for write file.
- When writing a switch statement for a const enum, DO NOT INCLUDE a default clause. This is to fully leverage the exhaustive linter.
- Use `utils.ParseMarkdownWithMetadata` for parsing Markdown with frontmatter.
