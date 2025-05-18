---
attach: always
---

# Project Conventions & Notes

- Always use English for codes, comments, and documentation.
- Always prefer immutable data mutation.

## Scripts

- `mise run lint-fix`: Run `golangci-lint` with the `--fix` option to automatically fix linting issues.
- `mise run fmt`: Format the code.
- `mise run test`: Run all tests.
  - This is just a wrapper for `gotestsum`, so add arguments if you want to run specific tests.

## Testing

- When you complete each task in an issue, and when you complete all tasks in an issue, you must test the entire project with `mise run test`.
- **If making changes that alter input/output behavior:** First, update tests to expect the new behavior *before* applying the code changes.
- **If making changes that do NOT alter input/output behavior:** Do not modify tests. Ensure all existing tests pass after applying the code changes.
