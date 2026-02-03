# GitScribe (gs)

```shell
           /$$   /$$                                  /$$ /$$
          |__/  | $$
  /$$$$$$  /$$ /$$$$$$   /$$$$$$$  /$$$$$$$  /$$$$$$  /$$| $$$$$$$   /$$$$$$ 
 /$$__  $$| $$|_  $$_/  /$$_____/ /$$_____/ /$$__  $$| $$| $$__  $$ /$$__  $$ 
| $$  \ $$| $$  | $$   |  $$$$$$ | $$      | $$  \__/| $$| $$  \ $$| $$$$$$$
| $$  | $$| $$  | $$ /$$\____  $$| $$      | $$      | $$| $$  | $$| $$_____/
|  $$$$$$$| $$  |  $$$$//$$$$$$$/|  $$$$$$$| $$      | $$| $$$$$$$/|  $$$$$$$
 \____  $$|__/   \___/ |_______/  \_______/|__/      |__/|_______/  \_______/
 /$$  \ $$
|  $$$$$$/
 \______/
```

**Your AI-powered multi-agent git assistant.**

GitScribe analyzes your staged changes and generates Conventional Commit messages using your preferred AI model. It manages multiple providers and keeps your credentials safe using your system's secure keyring.

## Features

-   **Multi-Agent Support**: Choose between OpenAI (GPT-4o), Anthropic (Claude 3.5), OpenCode (Zen), or Groq (Llama 3.3).
-   **Secure Key Storage**: All API keys are encrypted and stored in your **OS Keyring** (Keychain, Gnome Keyring, etc.), never in plain text.
-   **Interactive Selection**: Easily browse and enable models with a polished TUI.
-   **All-in-One Workflow**: Stage, commit, and push with a single command.

## Installation

### From GitHub Releases (Recommended)

#### Linux
1.  Download the binary from the [latest release](https://github.com/albuquerquesz/gitscribe/releases/latest).
2.  Extract and move to your path:
    ```shell
    tar -xzf gs_linux_amd64.tar.gz
    sudo mv gs /usr/local/bin/
    ```

#### Windows
1.  Download `gs_windows_amd64.tar.gz` from the [releases page](https://github.com/albuquerquesz/gitscribe/releases/latest).
2.  Extract `gs.exe` and add it to your system `PATH`.

### Using `go install`
```shell
go install github.com/albuquerquesz/gitscribe@latest
```

## Setup

Before using GitScribe for the first time, you need to enable at least one AI model.

### `gs models`
Browse the catalog and set up your API keys.
```shell
gs models
```
- Use arrow keys to select a **Provider**.
- Choose a **Model** and press Enter.
- Paste your API key when prompted.
- The last model you configure automatically becomes your **default**.

## Usage

### `gs cmt`
Stages files, generates a commit message using the default agent, and pushes.

**Standard usage (Stages all and generates message):**
```shell
gs cmt
```

**Stage specific files:**
```shell
gs cmt main.go internal/auth/
```

**Use a specific agent (overrides default):**
```shell
gs cmt --agent anthropic-claude-3-5-sonnet
```

**Custom message (skips AI):**
```shell
gs cmt -m "feat: add secure key storage"
```

### `gs auth status`
Check which providers have keys stored in the keyring.
```shell
gs auth status
```

### `gs update`
Keep your CLI up to date.
```shell
gs update
```

## Security

GitScribe takes security seriously. It uses the `zalando/go-keyring` library to interact with your operating system's native secret management. Your API keys are stored in an encrypted state and are only decrypted in memory when making a request to the AI provider.

---
Built with ❤️ using [Charmbracelet](https://charmbracelet.com/) tools.
