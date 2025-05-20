# ajisai

**Ajisai** is a simple preset manager for AI Coding Agents.

## Features

- **Interoperability 🤖** - Simply by writing rules and prompts in a single format, they are automatically deployed to the appropriate format and directory for each supported AI Coding Agent.
- **Reuse 📤** - You can import AI presets not only from local directories but also from remote Git repositories.
- **Gradual Introduction ⏩**: Enables adoption without conflicting with existing rules.

## Supported AI Coding Agents

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
- [ ] Roo Code
- [ ] Devin

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

## Getting Started

`ajisai` is controlled via CLI.

The default configuration file is `ajisai.yml` in the current directory, but a different file can be specified using the `--config` (or `-c`) flag.

> [!NOTE]
> `ajisai.yaml` will be supported soon.

### 1. Write Config

```yaml
workspace:
  imports:
    local_rules:
      type: local
      path: "./.ai"
      include:
      - default # Just HACK, will be documented soon.
  integrations:
    cursor:
      enabled: true
    github-copilot:
      enabled: true
    windsurf:
      enabled: true
```

### 2. Write your rules

Write your rules under `.ai/rules/**/*.md`.

Refer [Defining Preset](#defining-preset) for supported syntax.

### 3. Deploy your rules

Just run `ajisai apply`.

## Defining preset

### Rule File (`*.md`)

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

### Prompt File (`*.md`)

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

## Export your presets as a package

You can export your rules as a package and share via Git.

1. Write rules under `essential/rules/**/*.md` with [supported syntax](#defining-preset).
2. Place `ajisai.yml` at repository root.

      ```yaml
      # ajisai.yml in your org's rule repository
      package:
        exports:
          essential: # This means export rules and prompts below as `essential` preset.
            rules:
            - README.md
            - essential/rules/**/*.md
            prompts:
            - essential/prompts/**/*.md
      ```

3. exported `essential` preset can be included from other workspace.

      ```yaml
      # ajisai.yml in your workspace
      workspace:
        imports:
          org-essential:
            type: git
            repository: org-rules-repository-url
            include:
            - essential
        integrations:
          cursor:
            enabled: true
          # other integrations config...
      ```

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/sushichan044/ajisai).
