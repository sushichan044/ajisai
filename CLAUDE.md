# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ajisai is a CLI tool written in Go that manages AI agent configuration presets. It allows packaging and reusing rule and prompt configurations across multiple AI coding agents (Cursor, GitHub Copilot, Windsurf) and projects.

## Development Commands

### Building and Running

- **Run in development**: `mise run dev` or `go run main.go`
- **Build with goreleaser**: `mise run build-snapshot`
- **Apply AI rules**: `mise run apply` (runs `ajisai apply`)

### Testing and Quality

- **Run tests**: `mise run test` (uses gotestsum)
- **Run tests with coverage**: `mise run test-coverage`
- **Lint code**: `mise run lint` (golangci-lint)
- **Auto-fix lint issues**: `mise run lint-fix`
- **Format code**: `mise run fmt`

### Utilities

- **Clean generated files**: `mise run clean`
- **Clean cache**: `go run main.go clean` or `go run main.go clean --force`

## Architecture Overview

### Core Workflow

The main workflow follows: **Fetch → Load → Export**

1. **Fetch**: Retrieve packages from sources (local filesystem or Git repositories)
2. **Load**: Parse preset packages from cache, handling Markdown files with YAML frontmatter
3. **Export**: Convert domain objects to agent-specific formats and write to appropriate directories

### Key Components

**CLI Layer (`cmd/ajisai/`)**: Command-line interface using urfave/cli/v3

- Main commands: `apply`, `clean`, `import`, `doctor`
- Configuration loading and context management

**Engine Layer (`internal/engine/`)**: Core orchestration logic

- `ApplyPackage()`: Main workflow coordination
- `CleanOutputs()`: Remove generated integration files
- `CleanCache()`: Manage cached packages

**Domain Layer (`internal/domain/`)**: Business entities and interfaces

- `AgentPresetPackage`: Container for multiple presets
- `AgentPreset`: Contains rules and prompts
- `RuleItem`/`PromptItem`: Individual rule/prompt with metadata
- Attachment types: `always`, `glob`, `agent-requested`, `manual`

**Infrastructure Layer**:

- `internal/fetcher/`: Package retrieval (Local/Git fetchers)
- `internal/integration/`: AI agent output generation (Cursor, GitHub Copilot, Windsurf)
- `internal/bridge/`: Convert domain models to agent-specific formats
- `internal/loader/`: Parse packages from cache using glob patterns

**Configuration (`internal/config/`)**: YAML configuration management

- Supports `ajisai.yml`/`ajisai.yaml` files
- Three main sections: Settings, Package, Workspace

### AI Agent Integrations

Each integration defines:

- **File paths**: Where to write generated files
- **Extensions**: `.mdc` for Cursor rules, `.md` for others
- **Format**: How to serialize domain objects

**Generated file structure**:

```
.cursor/rules/ajisai/package-name/preset-name/rule-slug.mdc
.cursor/prompts/ajisai/package-name/preset-name/prompt-slug.md
```

### Testing Strategy

- Unit tests for each package with `_test.go` files
- Mock implementations in `utils/mocks/`
- Test data and fixtures included in test files
- Use testify/assert and go-mock for mocking

### Build and Version Management

- Uses goreleaser for releases
- Version embedding at build time (revision variable in main.go)
- mise.toml defines development tasks and tool versions
- CI/CD with GitHub Actions (lint and test jobs)

## Configuration Files

**Development**: `ajisai.yml` - workspace configuration for importing presets
**Package Definition**: `ajisai.yml` with `package.exports` - defines exportable presets
**Tool Management**: `mise.toml` - defines development tasks and dependencies

## Key Dependencies

- `github.com/urfave/cli/v3`: CLI framework
- `github.com/goccy/go-yaml`: YAML processing
- `github.com/adrg/frontmatter`: Markdown frontmatter parsing
- `github.com/bmatcuk/doublestar/v4`: Glob pattern matching
- `golang.org/x/sync/errgroup`: Concurrent processing
