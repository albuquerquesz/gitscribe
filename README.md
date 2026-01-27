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

**Your AI-powered git commit assistant.**

gitscribe analyzes your staged changes and generates conventional commit messages for you, making it easier to maintain a clean and informative git history.

## Features

-   **AI-Powered Commits**: Automatically generates commit messages based on your staged changes using AI.
-   **Conventional Commits**: Enforces the Conventional Commits specification for consistent and readable history.
-   **All-in-One Command**: Stage, commit, and push your changes with a single command.
-   **Automatic Updates**: Stay up-to-date with the latest version with built-in update notifications and command.

## Installation

You can install gitscribe in two ways:

### From GitHub Releases (Recommended for most users)

#### Linux

1.  Go to the [latest release page](https://github.com/albuquerquesz/gitscribe/releases/latest).
2.  Download the binary for your operating system and architecture (e.g., `gs_linux_amd64.tar.gz`).
3.  Extract the archive. You will find a binary file named `gs`.
    ```shell
    tar -xzf gs_linux_amd64.tar.gz
    ```
4.  Move the `gs` binary to a directory in your system's `PATH`. For example:
    ```shell
    sudo mv gs /usr/local/bin/
    ```
5.  Verify the installation by running:
    ```shell
    gs --version
    ```

#### Windows

1.  Go to the [latest release page](https://github.com/albuquerquesz/gitscribe/releases/latest).
2.  Download the binary for your operating system and architecture (e.g., `gs_windows_amd64.tar.gz`).
3.  Extract the archive. You will find a binary file named `gs.exe`. You can use a tool like 7-Zip or WinRAR, or the `tar` command in PowerShell or Command Prompt:
    ```shell
    tar -xzf gs_windows_amd64.tar.gz
    ```
4.  Move the `gs.exe` binary to a directory in your system's `PATH`. For example, you can create a folder `C:\Program Files\gitscribe` and add it to your `PATH`.
5.  Verify the installation by opening a new terminal and running:
    ```shell
    gs --version
    ```

### Using `go install` (For Go developers)

If you have Go installed, you can use `go install`:
```shell
go install github.com/albuquerquesz/gitscribe/cmd/gs@latest
```

## Usage

gitscribe is easy to use. Here are the available commands:

### `gs cmt`

This is the main command. It stages files, generates a commit message, commits the changes, and pushes them to the remote repository.

**Stage all files, generate a commit message, and push to the `main` branch:**
```shell
gs cmt
```

**Stage specific files:**
```shell
gs cmt file1.go file2.go 
  OR
gs cmt *.go
```

**Provide your own commit message (skips AI generation):**
```shell
gs cmt -m "feat: add new feature"
```

**Push to a different branch:**
```shell
gs cmt -b my-feature-branch
```

### `gs update`

Checks for a new version of gitscribe and prompts you to update if one is available.
```shell
gs update
```
gitscribe also checks for updates in the background and will notify you when a new version is available.

### `gs --version`

Prints the current version of gitscribe.
```shell
gs --version
```

