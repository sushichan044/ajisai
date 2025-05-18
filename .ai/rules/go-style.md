---
attach: glob
globs:
  - "**/*.go"
---

# Go Style Guide

- Use `errors.Is` for error comparison.
- Use `utils.AtomicWriteFile` for write file.
- When writing a switch statement for a const enum, DO NOT INCLUDE a default clause. This is to fully leverage the exhaustive linter.
