---
attach: always
---

# Project Conventions & Notes

- Always use English for project artifacts, comments, and documentation.
- Always prefer immutable data mutation.

## Testing

- When you complete each task in an issue, and when you complete all tasks in an issue, you must test the entire project with `go test`.
- **If making changes that alter input/output behavior:** First, update tests to expect the new behavior *before* applying the code changes.
- **If making changes that do NOT alter input/output behavior:** Do not modify tests. Ensure all existing tests pass after applying the code changes.

## Go Modules & Vendoring

- During local development where internal packages are imported, ensure a `replace` directive exists in @go.mod pointing to the local path (e.g., `replace github.com/sushichan044/aisync => ./`).
- This project uses Go Modules with vendoring. After running `go mod tidy` to update dependencies (especially test dependencies), always run `go mod vendor` to keep the `vendor/` directory synchronized.
