---
attach: always
---

# Project Conventions & Notes

- Always use English for codes, comments, and documentation.
- Always prefer immutable data mutation.

## Local Workflow

1. Lint / Format your changes by yourself:

    ```bash
    mise run lint-fix
    mise run fmt
    ```

2. Run Test by your self:

    ```bash
    mise run test
    ```

    You can run specific tests. e.g. `mise run test ./internal/config/...`

## Testing

- When you complete each task in an issue, and when you complete all tasks in an issue, you must test the entire project with `mise run test`.
- **If making changes that alter input/output behavior:** First, update tests to expect the new behavior *before* applying the code changes.
- **If making changes that do NOT alter input/output behavior:** Do not modify tests. Ensure all existing tests pass after applying the code changes.
