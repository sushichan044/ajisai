# aisync

<!-- TOC -->

- [aisync](#aisync)
  - [Features](#features)
  - [AI Coding Agents support status](#ai-coding-agents-support-status)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Commands](#commands)
  - [Defining preset](#defining-preset)
    - [Rule File (`*.md` in `rules/`)](#rule-file-md-in-rules)
    - [Prompt File (`*.md` in `prompts/`)](#prompt-file-md-in-prompts)
  - [Configuration](#configuration)
  - [Contributing](#contributing)

<!-- /TOC -->

> [!WARNING]
> This software is still under development. Be careful.

**aisync** is a command-line tool designed to manage and synchronize AI coding agent configuration presets (like rules and prompts) from various sources to different environments.

It helps you keep your AI assistant\'s behavior consistent across platforms or easily share and version control your configurations.

## Features

- Manage all your AI preset sources and destinations from a single `aisync.toml` file.
- Fetch presets from local directories or remote Git repositories (specific branches, tags, or commits).
- Export presets to formats and locations suitable for different AI coding agents (e.g., Cursor, GitHub Copilot).
- *(Upcoming)* `import` command to import presets from existing agent formats.
- *(Upcoming)* `doctor` command to validate preset directory structures and configurations.

## AI Coding Agents support status

- [x] GitHub Copilot in VSCode
  - Update VSCode to 1.100 or later
  - Use latest GitHub Copilot extension
  - **Make sure you set `chat.promptFiles` to true in project or user settings.**
  - `mode` and `tools` property in prompt file is not supported yet
  - Currently no support for using GitHub Copilot in other editors, IDEs.
- [x] Cursor
- [ ] Windsurf
- [ ] Cline
- [ ] Roo Code

## Installation

WIP

## Usage

`aisync` is controlled via command-line arguments and flags.

The default configuration file is `aisync.toml` in the current directory, but a different file can be specified using the `--config` (or `-c`) flag.

### Commands

- **`aisync apply`**: Fetches presets from all configured input sources, parses them, cleans the output directories, and then exports them to the configured output targets.

    ```bash
    aisync apply
    aisync apply --config /path/to/your/custom-config.toml
    ```

- **`aisync clean`**: Cleans the cache directory used by `aisync` to store fetched presets. By default, it cleans the cache specified in the configuration. The `--force` flag can be used to ensure cleaning even if errors occur during the process.

    ```bash
    aisync clean
    aisync clean --force
    ```

To see all available commands and options, use:

```bash
aisync --help
```

## Defining preset

A preset is a directory containing rules and prompts that `aisync` can manage. The expected directory structure for a preset is as follows:

```
your-preset-name/
├── rules/
│   ├── rule1.md
│   └── rule2.md
└── prompts/
    ├── prompt1.md
    └── prompt2.md
```

- **`rules/`**: This directory contains Markdown files defining rules. Each file represents a single rule.
- **`prompts/`**: This directory contains Markdown files defining prompts. Each file represents a single prompt.

The filename (without the `.md` extension) serves as the unique ID for that rule or prompt.

### Rule File (`*.md` in `rules/`)

Each rule Markdown file can have the following metadata in its frontmatter:

| Key           | Type    | Required | Description                                                                                                |
|---------------|---------|----------|------------------------------------------------------------------------------------------------------------|
| `attach` | String  | Yes       | Situation you want AI to read this rule. <br> Choose from `always`, `glob`, `agent-requested`, `manual`.  |
| `glob`      | Array  | Yes (when `attach` is `glob`)       | An array of glob patterns specifying which files this rule should apply to. <br> (e.g., `["**/*.go", "!**/*_test.go"]`). |
| `description` | String  | No       | A brief description of what the prompt is for.                                                                |

Example `rules/my-custom-rule.md`:

```markdown
---
attach: always
glob:
  - "**/*.go"
  - "!**/*_test.go"
---

This is the main content of the rule.
It describes the coding standard in detail...
```

### Prompt File (`*.md` in `prompts/`)

| Key           | Type    | Required | Description                                                                                                   |
|---------------|---------|----------|---------------------------------------------------------------------------------------------------------------|
| `description` | String  | No       | A brief description of what the prompt is for.                                                                |

Example `prompts/my-refactor-prompt.md`:

```markdown
---
description: A prompt to help refactor Go code for better readability.
---

Please refactor the following Go code to improve its readability and maintainability, keeping in mind our company's Go coding standards.
```

## Configuration

`aisync` is configured using a TOML file, by default named `aisync.toml`. Here is an example configuration:

```toml
[settings]
# Specifies the directory where `aisync` will store cached data of inputs.
cacheDir = "./.cache/aisync" # Optional: Defaults to `./.cache/aisync`.

# A namespace string that can be used by output targets to organize or prefix the imported presets.
# For example, aisync might place presets under `~/.cursor/prompts/<namespace>/` or `~/.cursor/rules/<namespace>/`.
namespace = "aisync"      # Optional: Defaults to `aisync`.

experimental = false # Optional: Defaults to false. Set to true to enable experimental features.

[inputs.local_rules]
# Input presets from a local directory.
type = "local"
path = "./.ai" # Required: Path to the directory containing presets

[inputs.my_ai_presets_example1]
# Input presets from a Git repository.
type = "git"
repository = "https://github.com/sushichan044/ai-presets.git" # Required: URL of the Git repository
revision = "main" # Optional: Specify a branch, tag, or commit SHA. Defaults to the repo's default branch.
directory = "my-company-requirements" # Optional: import from subdirectory.

[outputs.cursor]
# Export presets for the Cursor editor.
target = "cursor"  # Identifier for the output type

# set false to ignore this output.
enabled = true     # Optional, default: true

[outputs.github_copilot]
target = "github-copilot"
enabled = true
```

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/sushichan044/aisync).

(Further details on development setup, coding standards, and the contribution process can be added here.)
