# ajisai

[![ci](https://github.com/sushichan044/ajisai/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/sushichan044/ajisai/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sushichan044/ajisai.svg)](https://pkg.go.dev/github.com/sushichan044/ajisai)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/sushichan044/ajisai)

**Ajisai** is a simple preset manager for AI Coding Agents.

You can package rule and prompt configurations and reuse them across multiple projects.

- [ajisai](#ajisai)
  - [Features](#features)
  - [Supported AI Coding Agents](#supported-ai-coding-agents)
  - [Installation](#installation)
  - [Quick Start](#quick-start)
    - [1. Write Config](#1-write-config)
    - [2. Write your rules](#2-write-your-rules)
    - [3. Deploy your rules](#3-deploy-your-rules)
  - [User Guide](#user-guide)
    - [Defining and Importing Local Preset Packages](#defining-and-importing-local-preset-packages)
      - [1. Define Your Package's Exports](#1-define-your-packages-exports)
      - [2. Structure Your Rule and Prompt Files](#2-structure-your-rule-and-prompt-files)
      - [3. Import the Local Package into Your Workspace](#3-import-the-local-package-into-your-workspace)
    - [Sharing and Exporting Packages via Git](#sharing-and-exporting-packages-via-git)
    - [Import Preset Packages via Git](#import-preset-packages-via-git)
    - [Tip: Special `default` preset](#tip-special-default-preset)
  - [File Reference](#file-reference)
    - [Rule File (`*.md`)](#rule-file-md)
    - [Prompt File (`*.md`)](#prompt-file-md)
  - [Config Reference](#config-reference)
  - [Contributing](#contributing)

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
  <summary>go install (requires Go 1.21+) </summary>

```bash
go install github.com/sushichan044/ajisai/cmd/ajisai@latest
```

> [!WARNING]
> Because `ajisai` embeds its version information at **build time**, the **version** isn't displayed correctly when installed using `go install`.

</details>

<details>
  <summary>mise</summary>

```toml
# mise.toml
[tools]
"go:github.com/sushichan044/ajisai/cmd/ajisai" = "latest"
```

> [!WARNING]
> Because `ajisai` embeds its version information at **build time**, the **version** isn't displayed correctly when installed using `go install`.

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

Refer [Rule file - File Reference](#rule-file-md) for supported syntax and structure.

### 3. Deploy your rules

Just run `ajisai apply`.

## User Guide

In ajisai, instructions for AI Coding Agents are handled using the following units:

- **Preset**: A collection of specific Rules and reusable prompts.
- **Package**: A unit for exporting multiple presets.

When reusing packaged instructions, you specify the package to use and the presets to include from it.

### Defining and Importing Local Preset Packages

You can define reusable preset packages locally, for example, within a dedicated directory in your project (like `.ai/`) or in a separate local directory. This same package definition approach is fundamental, whether you intend to use the package only locally or later share it via Git.

#### 1. Define Your Package's Exports

Create an `ajisai.yml` or `ajisai.yaml` file in the root directory of your intended package (e.g., `<project root>/.ai/ajisai.yaml`). In this file, you define what presets your package will export using the `package.exports` field. Each key under `exports` becomes a named preset that can be imported.

   ```yaml
   # Example: <project root>/.ai/ajisai.yaml defining a package with an 'essential' preset
   package:
     exports:
       # 'essential' is the name of the preset being exported from this package.
       # Users will refer to this name when importing.
       essential:
         description: "Essential coding standards and prompts for the project." # Optional
         rules:
           # List of glob patterns for rule files relative to this ajisai.yaml
           - README.md # You can include markdown files directly as rules
           - essential/rules/**/*.md
         prompts:
           # List of glob patterns for prompt files relative to this ajisai.yaml
           - essential/prompts/**/*.md
       # You can define and export multiple presets from a single package file:
       # project-specific-utils:
       #   rules:
       #     - utils/rules/**/*.md
       #   prompts:
       #     - utils/prompts/**/*.md
   ```

#### 2. Structure Your Rule and Prompt Files

Organize your actual rule and prompt files according to the paths (glob patterns) you specified in the `package.exports` section. These paths are relative to the location of this package `ajisai.yaml` file.

   For the `essential` preset example above, the directory structure within `.ai/` might look like this:

   ```plaintext
   <project root>
   └── .ai/
       ├── essential/
       │   ├── rules/
       │   │   ├── common-guidelines.md
       │   │   └── go-specific.md      # Included by essential/rules/**/*.md
       │   └── prompts/
       │       └── refactor-prompt.md  # Included by essential/prompts/**/*.md
       ├── README.md                   # Directly included as a rule
       └── ajisai.yaml                 # The package definition file itself
   ```

   Any rule file created or matching the glob patterns (e.g., a new file in `.ai/essential/rules/`) will automatically become part of the `essential` preset. Refer to the [File Reference](#file-reference) for the specific format and frontmatter expected in rule and prompt files.

#### 3. Import the Local Package into Your Workspace

To use this locally defined package in your main project (or any other project that can access this path), modify your primary `ajisai.yml` (usually at the project root) to import it using `type: local`.

   ```yaml
   # <project root>/ajisai.yaml (Main workspace configuration)
   workspace:
     imports:
       # 'my_local_essentials' is an arbitrary name for this import instance in your workspace.
       my_local_essentials:
         type: local
         path: ./.ai  # Path to the directory containing the package's ajisai.yaml
         include:
           - essential # Specify the name of the preset(s) to import from that package.
     # ... other workspace configurations like integrations
     integrations:
       cursor:
         enabled: true
       # ...
   ```

   This setup allows you to manage and version control your shared AI instructions within a subdirectory of your project or a dedicated local repository.

### Sharing and Exporting Packages via Git

To share your presets as a package via Git, allowing others (or yourself in different projects) to reuse them:

1. Create an `ajisai.yml` or `ajisai.yaml` at repository root.

2. Define exported preset in config file and place your preset content in the same way as [Defining and Importing Local Preset Packages](#defining-and-importing-local-preset-packages) section.

3. Commit and push.

Your package is now ready to be imported by others using its Git repository URL.

### Import Preset Packages via Git

For example, to import the `essential` preset from a package shared via Git (as defined in the "[Sharing and Exporting Packages via Git](#sharing-and-exporting-packages-via-git)" guide), add the following configuration to the `ajisai.yml` in the project root of the importing workspace:

> [!NOTE]
> You need to have access to the repository where the package definitions are stored.

```yaml
# ajisai.yml in your workspace
workspace:
  imports:
    org-essential: # you can specify any name to identify imported preset packages.
      type: git
      repository: your-preset-package-repository-url # URL of the Git repository
      include:
      - essential # deploy `essential` preset from that package.

  # In `integrations`, you specify the AI Coding Agent that will actually utilize the imported preset package.
  integrations:
    cursor:
      enabled: true
    github-copilot:
      enabled: true
    windsurf:
      enabled: true
```

### Tip: Special `default` preset

If you do not have an `ajisai.yml` or `ajisai.yaml` file in your package root (e.g., a simple Git repository with just rules/prompts in a conventional structure), but your project adheres to a special directory structure as shown below, you can specify `default` in the `include` setting to have this structure recognized as a preset.

- Write rules at `<package root>/rules/**/*.md`
- Write prompts at `<package root>/prompts/**/*.md`

So you can import this to your workspace with:

```yaml
workspace:
  imports:
    org-default:
      type: git
      repository: org-rules-repo-url # A repo with files in <root>/rules/ and/or <root>/prompts/
      include:
      - default # This 'default' refers to the special auto-detected preset
    local-default:
      type: local
      path: "./.ai"
      include:
      - default
```

## File Reference

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

## Config Reference

```yaml
# This file (`ajisai.yml` or `ajisai.yaml`) can define EITHER a package OR a workspace, but not both.

# To define a re-usable package (typically placed at the root of a package repository or a dedicated local directory):
package:
  name: "sushichan044/example" # Optional: Package name. Currently has no major effect but can be used for identification.
  exports: # Define presets exported by this package.
    essential: # This is the preset name, e.g., 'essential'.
      description: "Core set of rules and prompts." # Optional
      rules: # Glob patterns for rule files, relative to this ajisai.yml
      - README.md
      - essential/rules/**/*.md
      prompts: # Glob patterns for prompt files, relative to this ajisai.yml
      - essential/prompts/**/*.md
    # another-preset:
    #   ...

# To define a workspace configuration (typically placed at your project root):
workspace:
  # Defines the preset packages to be used in this workspace.
  imports:
    local_rules: # Arbitrary identifier for this import source
      type: local
      path: "./.ai" # Path to the directory containing the package's ajisai.yml
      include: # List of preset names to import from that package
      - default # e.g., 'default' if the local package exports a 'default' preset or uses the special default structure
    remote_rules:
      type: git
      repository: https://github.com/sushichan044/ai-presets.git
      include:
      - example1 # Name of a preset exported by the package in the Git repository
  # Defines which AI Coding Agent integrations will utilize the imported presets.
  integrations:
    cursor:
      enabled: true # Set to true to deploy applicable presets for Cursor
    github-copilot:
      enabled: true
    windsurf:
      enabled: true

settings:
  # Specifies the directory where ajisai temporarily caches imported packages.
  cacheDir: "./.cache/ajisai" # default: ./.cache/ajisai

  # Sets the namespace that ajisai uses when deploying imports.
  # This helps avoid conflicts if multiple tools write to similar paths.
  # For example, if the namespace is `ajisai`, Cursor Rules are deployed to `.cursor/rules/ajisai/**/*.mdc`.
  namespace: ajisai # default: ajisai

  # Whether to enable experimental features.
  experimental: false # default: false
```

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues on the [GitHub repository](https://github.com/sushichan044/ajisai).
