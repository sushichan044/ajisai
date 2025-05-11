# aisync

> [!WARNING]
> This software is still under development. Be careful.

**aisync** is a command-line tool designed to manage and synchronize AI coding agent configuration presets (like rules and prompts) from various sources to different environments.

It helps you keep your AI assistant\'s behavior consistent across platforms or easily share and version control your configurations.

## Features

* **Centralized Configuration**: Manage all your AI preset sources and destinations from a single `aisync.toml` file.
* **Multiple Input Sources**:
  * Fetch presets from local directories.
  * Fetch presets from Git repositories (specific branches, tags, or commits).
* **Flexible Output Targets**: Export presets to formats and locations suitable for different AI coding agents (e.g., Cursor, GitHub Copilot).
* **Apply Presets**: Cleans output directories and applies the latest fetched and parsed presets.
* **Cache Management**: Fetched presets are cached locally, and the cache can be cleaned using the `clean` command.
* **Cross-Platform**: Builds available for Linux, macOS, and Windows.
* *(Upcoming)* `import` command to import presets from existing agent formats.
* *(Upcoming)* `doctor` command to validate preset directory structures and configurations.

## AI Coding Agents support status

* [x] GitHub Copilot
  * Make sure you set `chat.promptFiles` to true.
* [x] Cursor
* [ ] Windsurf
* [ ] Cline
* [ ] Roo Code

## Installation

WIP

## Usage

`aisync` is controlled via command-line arguments and flags.

The default configuration file is `aisync.toml` in the current directory, but a different file can be specified using the `--config` (or `-c`) flag.

### General Synopsis

```bash
aisync [global options] command [command options] [arguments...]
```

### Commands

* **`aisync apply`**: Fetches presets from all configured input sources, parses them, cleans the output directories, and then exports them to the configured output targets.

    ```bash
    aisync apply
    aisync apply --config /path/to/your/custom-config.toml
    ```

* **`aisync clean`**: Cleans the cache directory used by `aisync` to store fetched presets. By default, it cleans the cache specified in the configuration. The `--force` flag can be used to ensure cleaning even if errors occur during the process.

    ```bash
    aisync clean
    aisync clean --force
    ```

To see all available commands and options, use:

```bash
aisync --help
```

## Configuration

`aisync` is configured using a TOML file, by default named `aisync.toml`. Here is an example configuration:

```toml
[global]
# Specifies the directory where `aisync` will store cached data of inputs.
cacheDir = ".cache/aisync" # Optional: Defaults to `.cache/aisync/`.
# A namespace string that can be used by output targets to organize or prefix the imported presets.
# For example, aisync might place presets under `~/.cursor/prompts/<namespace>/` or `~/.cursor/rules/<namespace>/`.
namespace = "aisync"      # Optional: Defaults to `aisync`.

# --- Input Sources ---
# Define where your AI presets (rules, prompts) come from.

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

# --- Output Targets ---
# Define where the processed presets should be exported.

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
