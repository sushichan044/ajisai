# ajisai

**Ajisai** is a simple preset manager for AI Coding Agents.

## Features

- **Interoperability 🤖** - Simply by writing rules and prompts in a single format, they are automatically deployed to the appropriate format and directory for each supported AI Coding Agent.
- **Reuse 📤** - You can import AI presets not only from local directories but also from remote Git repositories.
- **Gradual Introduction ⏩**: Enables adoption without conflicting with existing rules.

<!-- ## Supported AI Coding Agents

- [x] GitHub Copilot in VSCode
  - Update VSCode to 1.100 or later
  - Use latest GitHub Copilot extension
  - **Make sure you set `chat.promptFiles` to true in project or user settings.**
  - `mode` and `tools` property in prompt file is not supported yet
  - Currently no support for using GitHub Copilot in other editors, IDEs.
- [x] Cursor
- [x] Windsurf
  - Update Windsurf to Wave 8 or later
- [ ] Cline
- [ ] Roo Code -->

## Installation

<details>
  <summary>homebrew</summary>

```bash
brew install sushichan044/tap/ajisai
```

</details>

<details>
  <summary>go install</summary>

```bash
go install github.com/sushichan044/ajisai@latest
```

</details>

<details>
  <summary>manual</summary>

Download the latest release from the [GitHub releases page](https://github.com/sushichan044/ajisai/releases).

</details>

## Usage

`ajisai` is controlled via CLI.

The default configuration file is `ajisai.json` in the current directory, but a different file can be specified using the `--config` (or `-c`) flag.

### Commands

- **`ajisai apply`**: Fetch all imported packages and deploy to integrated agents.

## Defining preset

A preset is a directory containing rules and prompts that `ajisai` can manage. The expected directory structure for a preset is as follows:

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
| `globs`      | Array  | Yes <br> (when `attach` is `glob`) <br> | An array of glob patterns specifying which files this rule should apply to. <br> (e.g., `["**/*.go", "!**/*_test.go"]`). |
| `description` | String  | Yes <br> (when `attach` is `agent-requested`) <br> | A brief description of what the prompt is for.                                                                |

Example `rules/my-custom-rule.md`:

```markdown
---
attach: always
globs:
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

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/sushichan044/ajisai).
