# GitScribe (gs)

```
           /$$   /$$                                  /$$ /$$                
          |__/  | $$                                 |__/| $$                
  /$$$$$$  /$$ /$$$$$$   /$$$$$$$  /$$$$$$$  /$$$$$$  /$$| $$$$$$$   /$$$$$$ 
 /$$__  $$| $$|_  $$_/  /$$_____/ /$$_____/ /$$__  $$| $$| $$__  $$ /$$__  $$
| $$  \ $$| $$  | $$   |  $$$$$$ | $$      | $$  \__/| $$| $$  \ $$| $$$$$$$$
| $$  | $$| $$  | $$ /$$\____  $$| $$      | $$      | $$| $$  | $$| $$_____/
|  $$$$$$$| $$  |  $$$$//$$$$$$$/|  $$$$$$$| $$      | $$| $$$$$$$/|  $$$$$$$
 \____  $$|__/   \___/ |_______/  \_______/|__/      |__/|_______/  \_______/
 /$$  \ $$                                                                    
|  $$$$$$/                                                                    
 \______/                                                                    
```

**Your AI-powered multi-agent git assistant with intelligent context management.**

GitScribe analyzes your staged changes and generates Conventional Commit messages using your preferred AI model. It manages multiple providers, keeps your credentials safe using your system's secure keyring, and now includes a powerful context system for project-specific instructions.

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
  - [`gs commit` - AI-Powered Commits](#gs-commit)
  - [`gs pr` - Pull Request Creation](#gs-pr)
  - [`gs context` - Context Management](#gs-context)
  - [`gs agent` - Agent Management](#gs-agent)
  - [`gs models` - Model Browser](#gs-models)
  - [Other Commands](#other-commands)
- [Context System](#context-system)
- [Security](#security)
- [Configuration](#configuration)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

---

## Features

- **Multi-Agent Support**: Choose between OpenAI (GPT-4o), Anthropic (Claude), Groq (Llama), OpenCode (Kimi), and more
- **Context System**: Add project-specific instructions (up to 3 per project) to guide AI commit generation
- **PR Creation**: Auto-detect GitHub/GitLab and create PRs with AI-generated titles and descriptions
- **Secure Key Storage**: All API keys encrypted in your **OS Keyring** (Keychain, GNOME Keyring, etc.)
- **Interactive UI**: Edit messages, view contexts, cancel or confirm - all with keyboard shortcuts
- **Visual Feedback**: Elegant spinners and styled output using Charmbracelet tools
- **All-in-One Workflow**: Stage, commit, push, and create PRs with a single tool

---

## Installation

### From GitHub Releases (Recommended)

#### Linux
```shell
# Download latest release
curl -L -o gs.tar.gz https://github.com/albuquerquesz/gitscribe/releases/latest/download/gs_linux_amd64.tar.gz

# Extract and install
tar -xzf gs.tar.gz
sudo mv gs /usr/local/bin/
rm gs.tar.gz
```

#### macOS
```shell
# Download latest release
curl -L -o gs.tar.gz https://github.com/albuquerquesz/gitscribe/releases/latest/download/gs_darwin_amd64.tar.gz

# Extract and install
tar -xzf gs.tar.gz
sudo mv gs /usr/local/bin/
rm gs.tar.gz
```

#### Windows
1. Download `gs_windows_amd64.tar.gz` from [releases page](https://github.com/albuquerquesz/gitscribe/releases/latest)
2. Extract `gs.exe` using 7-Zip or similar
3. Add to your system `PATH`

### Using `go install`
```shell
go install github.com/albuquerquesz/gitscribe@latest
```

### Using Homebrew (macOS/Linux)
```shell
brew tap albuquerquesz/gitscribe
brew install gs
```

---

## Quick Start

### 1. Initialize GitScribe
```shell
gs init
```
This will check for existing configurations and guide you through setup.

### 2. Configure Your First AI Model
```shell
gs models
```
- Select a provider (Anthropic, OpenAI, Groq, etc.)
- Choose a model
- Enter your API key (securely stored in OS keyring)

### 3. Make Your First Commit
```shell
gs commit
```
Or use the shorter alias:
```shell
gs cmt
```

### 4. (Optional) Add Project Context
```shell
gs ctx add "This is a React TypeScript project using Redux"
gs ctx add "Follow conventional commits with emoji prefixes"
```

---

## Commands

### `gs commit` (alias: `gs cmt`)

AI-powered git add, commit, and push workflow.

**Usage:**
```shell
# Stage all changes, generate commit with AI, commit and push
gs commit

# Stage specific files only
gs commit main.go internal/auth/

# Use custom message (skip AI generation)
gs commit -m "feat: add secure key storage"

# Use specific agent for this commit only
gs commit --agent claude-sonnet

# Push to specific branch
gs commit -b feature-branch
```

**Interactive Prompt:**
After generating the commit message, you'll see:
```
feat: implement user authentication

- Add JWT token validation
- Create auth middleware
- Update user model

[E] Edit  [C] Contexts  [ESC] Cancel  [↵] Continue
```

**Keyboard Shortcuts:**
- **E** - Edit the commit message inline
- **C** - View active contexts for this project
- **ESC** - Cancel the commit
- **Enter** - Proceed with the commit and push

---

### `gs pr`

Create pull requests on GitHub or GitLab with AI-generated title and description.

**Features:**
- Auto-detects provider (GitHub/GitLab) from remote URL
- Generates title and body from commit history
- Pushes branch automatically if needed
- Validates commits exist between branches

**Requirements:**
- [GitHub CLI (`gh`)](https://cli.github.com/) for GitHub repos
- [GitLab CLI (`glab`)](https://glab.readthedocs.io/) for GitLab repos

**Usage:**
```shell
# Generate PR title/body with AI and create PR
gs pr

# Create PR with custom title (body AI-generated)
gs pr -t "feat: implement new feature"

# Create PR with custom title and body
gs pr -t "feat: auth" -b "Detailed description here"

# Create draft PR targeting specific branch
gs pr --target develop --draft

# Target different branch
gs pr --target staging
```

**Workflow:**
1. Checks for uncommitted changes (warns if any)
2. Detects Git provider from remote URL
3. Verifies CLI tool is installed
4. Pushes current branch
5. Generates PR title/body from last 20 commits
6. Shows interactive prompt (similar to commit)
7. Creates PR using provider CLI

---

### `gs context` (alias: `gs ctx`)

Manage project-specific contexts to guide AI commit generation.

**What are Contexts?**
Contexts are instructions or information about your project that help the AI generate better commit messages. They're stored per git repository and included in every AI prompt.

**Limit:** Maximum 3 contexts per project (FIFO - First In, First Out)

#### `gs ctx add [contexto]`

Add a context string for the current project.

```shell
# Add tech stack context
gs ctx add "This is a Go project using Chi router and PostgreSQL"

# Add conventions context
gs ctx add "Use conventional commits with scope: feat(api):, fix(auth):, etc."

# Add team guidelines
gs ctx add "Always reference issue numbers: Fixes #123"
```

**Limit Check:**
```shell
gs ctx add "Fourth context"
# Error: Limite de 3 contextos atingido para este projeto
# Use 'gs ctx remove' para remover um contexto existente
```

#### `gs ctx list`

List all contexts for the current project.

```shell
gs ctx list
```

**Output:**
```
Contextos para /home/user/myproject (2/3):
1. This is a Go project using Chi router and PostgreSQL
2. Use conventional commits with scope: feat(api):, fix(auth):, etc.
```

#### `gs ctx remove` (alias: `gs ctx rm`)

Interactively remove a context.

```shell
gs ctx remove
```

**Interactive Flow:**
```
Contextos disponíveis:
1. This is a Go project using Chi router and PostgreSQL
2. Use conventional commits with scope

Qual deseja remover? (1-2): 1
✓ Contexto removido
```

Press `ESC` or enter invalid number to cancel.

#### How Contexts Affect AI Prompts

When you have contexts configured, the AI receives:

```
Contextos adicionais do projeto:
- This is a Go project using Chi router and PostgreSQL
- Use conventional commits with scope: feat(api):, fix(auth):, etc.

Analise o diff abaixo considerando os contextos acima:

[git diff content here]
```

This results in more relevant and project-appropriate commit messages.

---

### `gs agent`

Manage AI agent profiles.

#### `gs agent list`

List all configured agents.

```shell
gs agent list
```

**Output:**
```
★ 🟢 claude-sonnet       Provider: anthropic    Model: claude-3-5-sonnet    API Key: ✅
  🟢 groq-llama          Provider: groq          Model: llama-3.3-70b       API Key: ✅
  🔴 old-openai          Provider: openai        Model: gpt-4                API Key: ❌
```

**Legend:**
- `★` = Default agent
- `🟢`/`🔴` = Enabled/Disabled
- `✅`/`❌` = API key configured/not configured

#### `gs agent add`

Add a new agent profile.

```shell
# Add Groq agent
gs agent add -n my-groq -p groq -m llama-3.3-70b-versatile

# Add with API key inline
gs agent add -n production -p anthropic -m claude-3-opus -k sk-ant-xxx

# Add custom OpenAI-compatible endpoint
gs agent add -n local-ollama -p ollama -m llama2 --base-url http://localhost:11434/v1
```

**Flags:**
- `-n, --name`: Agent profile name (required)
- `-p, --provider`: Provider name (required)
  - Options: `anthropic`, `openai`, `groq`, `opencode`, `gemini`, `openrouter`, `ollama`
- `-m, --model`: Model name (required)
- `-k, --key`: API key (optional, prompts securely if not provided)
- `--base-url`: Custom base URL for custom endpoints

#### `gs agent remove [name]`

Remove an agent profile.

```shell
gs agent remove old-agent
```

#### `gs agent set-key [name]`

Update API key for an existing agent.

```shell
gs agent set-key my-agent
# Enter new API key securely
```

---

### `gs models`

Browse and enable AI models interactively.

```shell
gs models
```

**Interactive TUI:**
1. Select provider from list
2. Choose model from available options
3. Enter API key (masked input)
4. Model is enabled and set as default

**Features:**
- Visual model browser with descriptions
- Automatic API key validation
- Secure key storage

---

### Other Commands

#### `gs init`

Initialize GitScribe configuration.

```shell
gs init
```

Checks for:
- Existing configuration
- OpenCode authentication (`~/.local/share/opencode/auth.json`)
- Git repository setup

#### `gs update`

Self-update to the latest version.

```shell
gs update
```

Checks GitHub releases, shows changelog, and updates the binary automatically.

---

## Context System

The Context System is one of GitScribe's most powerful features. It allows you to provide project-specific instructions that guide the AI in generating better commit messages.

### Use Cases

**1. Tech Stack Information:**
```shell
gs ctx add "React 18 + TypeScript + Redux Toolkit project"
```

**2. Commit Conventions:**
```shell
gs ctx add "Use Angular commit convention: type(scope): subject"
gs ctx add "Available scopes: auth, api, ui, db, deps"
```

**3. Team Guidelines:**
```shell
gs ctx add "Reference Jira tickets in commits: PROJ-123"
gs ctx add "Mark breaking changes with BREAKING CHANGE: in body"
```

**4. Project Specifics:**
```shell
gs ctx add "This is a microservices architecture - mention service name"
gs ctx add "Database migrations should be explicitly mentioned"
```

### Storage

Contexts are stored in:
```
~/.config/gitscribe/contexts.json
```

Format:
```json
{
  "contexts": {
    "/home/user/projects/myapp": [
      {
        "text": "React TypeScript project",
        "created_at": "2026-02-04T10:00:00Z"
      }
    ]
  }
}
```

### Limits

- **3 contexts maximum** per git repository
- **FIFO ordering** - first added is first in the list
- **Path-based** - tied to git repository root (`git rev-parse --show-toplevel`)

### Viewing During Commit

During the commit flow, press `C` to see active contexts:
```
feat: implement user authentication

[E] Edit  [C] Contexts  [ESC] Cancel  [↵] Continue

[Press C]

Contextos ativos:
  • React 18 + TypeScript + Redux Toolkit project
  • Use Angular commit convention
```

---

## Security

### API Key Storage

GitScribe uses your **OS Native Keyring** for secure API key storage:

- **macOS**: Keychain
- **Linux**: Secret Service API / GNOME Keyring / KWallet
- **Windows**: Windows Credential Manager

Keys are:
- ✅ Encrypted at rest by the OS
- ✅ Never stored in plain text files
- ✅ Wiped from memory after use
- ✅ Accessible only to your user account

### Key Resolution Priority

When an API key is needed, GitScribe tries in order:

1. **Agent-specific keyring entry** (most specific)
2. **Generic keyring entry** (from config)
3. **Environment variables**:
   - `OPENAI_API_KEY`
   - `ANTHROPIC_API_KEY`
   - `GROQ_API_KEY`
   - `OPENCODE_API_KEY`
   - `GOOGLE_API_KEY`
   - `OPENROUTER_API_KEY`
4. **OpenCode auth file** (`~/.local/share/opencode/auth.json`)

### Secure Input

- API key prompts use password masking
- No sensitive data in logs
- Secure memory wiping after use

### File Permissions

```
~/.config/gitscribe/     (drwxr-xr-x)
├── config.yaml          (-rw-------)  # Owner only
├── contexts.json        (-rw-r--r--)  # Owner write, all read
└── ...
```

---

## Configuration

### Config File Location

```
~/.config/gitscribe/config.yaml
```

### Example Configuration

```yaml
version: "1.0"
global:
  default_agent: "claude-sonnet"
  auto_select: true
  request_timeout_seconds: 30
  max_retries: 3

agents:
  - name: "claude-sonnet"
    provider: "anthropic"
    model: "claude-3-5-sonnet-20241022"
    temperature: 0.7
    max_tokens: 4096
    timeout_seconds: 30
    enabled: true
    priority: 1
    keyring_key: "agent:claude-sonnet:api-key"

  - name: "groq-fast"
    provider: "groq"
    model: "llama-3.1-8b-instant"
    temperature: 0.5
    max_tokens: 2048
    enabled: true
    priority: 2
    keyring_key: "agent:groq-fast:api-key"

routing:
  - name: "quick-commits"
    agent_profile: "groq-fast"
    conditions: ["token_count < 1000"]
    priority: 1
```

### Supported Providers

| Provider | Models | Authentication |
|----------|--------|----------------|
| **Anthropic** | Claude 3.5 Sonnet, Claude 3.5 Haiku | API Key |
| **OpenAI** | GPT-4o, GPT-4o Mini, GPT-4 | Bearer Token |
| **Groq** | Llama 3.3 70B, Llama 3.1 8B | Bearer Token |
| **OpenCode** | Kimi 2.5, Mini Pickle, GLM | API Key |
| **Gemini** | Gemini 1.5 Pro, Gemini 1.5 Flash | API Key |
| **OpenRouter** | Various models | Bearer Token |
| **Ollama** | Local models (Llama2, Mistral, etc.) | None (local) |

---

## Examples

### Example 1: Basic Workflow

```shell
# Initialize
gs init

# Configure first model
gs models
# Select: Groq → Llama 3.3 70B → Enter API key

# Make changes
echo "console.log('hello')" > app.js

# Stage, commit with AI, push
gs commit
# Review message → Press Enter → Done!
```

### Example 2: Project with Contexts

```shell
# Setup project
cd my-react-project
git init

# Add contexts
gs ctx add "React 18 + TypeScript project with Redux Toolkit"
gs ctx add "Use conventional commits: feat:, fix:, docs:, style:, refactor:, test:, chore:"
gs ctx add "Available scopes: components, hooks, store, api, auth"

# Work on feature
gs commit -b feature/new-component
# AI generates: "feat(components): add UserProfile card with avatar"

# Create PR
gs pr
# AI generates PR title and description from commits
```

### Example 3: Multi-Agent Setup

```shell
# Add multiple agents
gs agent add -n claude -p anthropic -m claude-3-5-sonnet-20241022
gs agent add -n fast -p groq -m llama-3.1-8b-instant

# Use specific agent for complex refactoring
gs commit --agent claude

# Use fast agent for quick fixes
gs commit --agent fast -m "fix: typo in readme"

# List agents
gs agent list
```

### Example 4: PR Workflow

```shell
# Create feature branch
git checkout -b feature/auth-improvements

# Make several commits
gs commit -m "feat(auth): add JWT refresh token logic"
gs commit -m "feat(auth): implement token rotation"
gs commit -m "test(auth): add tests for token refresh"

# Push and create PR
gs pr --target main

# Or with custom title
gs pr -t "feat: implement secure token refresh mechanism"
```

### Example 5: Context Management Workflow

```shell
# Start new project
cd ~/projects/api-service
gs init

# Add contexts over time
gs ctx add "Go microservice using Chi router and PostgreSQL"
gs ctx add "Follow DDD patterns: repository, service, handler layers"
gs ctx add "Always include database migration notes in commits"

# Check contexts
gs ctx list
# Output: 3/3 contexts

# Try to add fourth (fails)
gs ctx add "Use structured logging with zap"
# Error: Limite de 3 contextos atingido

# Remove one
gs ctx remove
# Select: 2 (Follow DDD patterns...)

# Now can add new one
gs ctx add "Use structured logging with zap"
```

---

## Troubleshooting

### "No commits between main and <branch>"

**Cause:** Branch hasn't been pushed or has no commits

**Solution:**
```shell
git push origin <branch>
# Then retry
gs pr
```

### "gh CLI is not installed"

**Cause:** GitHub CLI not found in PATH

**Solution:**
```shell
# macOS
brew install gh

# Linux
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
sudo apt update && sudo apt install gh

# Or download from https://cli.github.com/
```

### "Limite de 3 contextos atingido"

**Cause:** Maximum contexts per project reached

**Solution:**
```shell
gs ctx remove
# Remove one context before adding new
```

### API Key not found

**Cause:** Key not in keyring or environment

**Solution:**
```shell
# Option 1: Set via agent command
gs agent set-key my-agent

# Option 2: Use environment variable
export GROQ_API_KEY="gsk_xxx"

# Option 3: Reconfigure model
gs models
```

### Commit cancelled shows error

**Behavior:** When pressing ESC during commit, previously showed "Error: commit cancelled"

**Current Behavior:** Now shows clean message without error:
```
ℹ Commit cancelled
```

### Spinner not showing

**Cause:** Terminal might not support ANSI codes

**Solution:** Try with `--accessible` flag (if available) or check terminal compatibility.

### Contexts not appearing in AI output

**Check:**
1. Are contexts added for current project? `gs ctx list`
2. Is current directory inside git repo? `git rev-parse --show-toplevel`
3. Try pressing `C` during commit to verify contexts are loaded

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`gs commit -m 'feat: add some amazing feature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request (`gs pr`)

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Built with [Charmbracelet](https://charmbracelet.com/) tools (Bubble Tea, Lipgloss, Huh)
- Multi-agent architecture inspired by modern LLM routing patterns
- Thanks to all contributors and users!

---

**Built with ❤️ and lots of ☕**
