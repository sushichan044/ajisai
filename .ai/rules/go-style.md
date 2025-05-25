---
attach: glob
globs:
  - "**/*.go"
---

# Local Rules for Go

## Style Guide

- Use `errors.Is` for error comparison.
- In test runner, use `require.NoError` for ensure operation success.
- When writing a switch statement for a const enum, DO NOT INCLUDE a default clause. This is to fully leverage the exhaustive linter.

## Utility Guide

- Use `utils.AtomicWriteFile` for write file.
- Use `utils.ParseMarkdownWithMetadata` for parsing Markdown with frontmatter.

## Testing Guide

- Use `package <package_name>_test` in test files for scope isolation.
