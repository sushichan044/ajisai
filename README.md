# ajisai

**Ajisai** is a simple preset manager for AI Coding Agents.

You can package rule and prompt configurations and reuse them across multiple projects.

<!-- TOC -->

- [ajisai](#ajisai)
  - [Features](#features)
  - [Supported AI Coding Agents](#supported-ai-coding-agents)
  - [Installation](#installation)
  - [Quick Start](#quick-start)
    - [1. Write Config](#1-write-config)
    - [2. Write your rules](#2-write-your-rules)
    - [3. Deploy your rules](#3-deploy-your-rules)
  - [Defining preset](#defining-preset)
    - [Rule File (`*.md`)](#rule-file-md)
    - [Prompt File (`*.md`)](#prompt-file-md)
  - [Export your presets as a package](#export-your-presets-as-a-package)
    - [Special `default` preset](#special-default-preset)
  - [Config Reference](#config-reference)
  - [Contributing](#contributing)

<!-- /TOC -->

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
- [x] Devin (Maybe partial support)
  - Devin can pull rules from the Cursor format, so enabling Cursor integration and run `ajisai apply` in Devin's environment would be effective.
    - <https://docs.devin.ai/onboard-devin/knowledge-onboarding#knowledge-101>

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

## Quick Start

`ajisai` is controlled via CLI.

The default configuration file is `ajisai.yml` or `ajisai.yaml` in the current directory, but a different file can be specified using the `--config` (or `-c`) flag.

### 1. Write Config

```yaml
workspace:
  imports:
    local_rules:
      type: local
      path: "./.ai"
      include:
      - default # See Special `default` preset docs for details
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

### Special `default` preset

For both Local Import and Git Import, if you do not include an `ajisai.yml` file and instead arrange your files in the structure shown below, a special `default` preset will be automatically recognized:

- Write rules at `<package root>/rules/**/*.md`
- Write prompts at `<package root>/rules/**/*.md`

So you can import this to your workspace with:

```yaml
workspace:
  imports:
    org-default:
      type: git
      repository: org-rules-repo-url
      include:
      - default
```

## Config Reference

```yaml
# Defines reusable preset packages that can be referenced from other workspaces.
package:
  name: "sushichan044/example" # Package name. currently has no effect.
  exports: # Define exported presets.
    essential: # This means export rules and prompts below as `essential` preset.
      rules:
      - README.md
      - essential/rules/**/*.md
      prompts:
      - essential/prompts/**/*.md

workspace:
  # Defines the preset packages to be used in this workspace.
  imports:
    local_rules:
      type: local
      path: "./.ai"
      include:
      - default
    remote_rules:
      type: git
      repository: https://github.com/sushichan044/ai-presets.git
      include:
      - example1
  # Defines the AI Coding Agent that deploys preset packages in this workspace.
  integrations:
    cursor:
      enabled: true
    github-copilot:
      enabled: true
    windsurf:
      enabled: true

settings:
  # Specifies the directory where ajisai temporarily caches the packages it imports.
  cacheDir: "./.cache/ajisai" # default: ./.cache/ajisai

  # Sets the namespace that ajisai uses when deploying imports.
  # For example, if the default namespace is `ajisai`, Cursor Rules are deployed to `.cursor/rules/ajisai/**/*.mdc`.
  namespace: ajisai # default: ajisai

  # Whether to enable experimental features.
  experimental: false # default: false
```

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/sushichan044/ajisai).
