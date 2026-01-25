# Roadmap

## Core Features
- [x] AI-powered commit message generation (Groq/Llama).
- [x] Secure API Key storage using system Keyring.
- [x] Interactive configuration via CLI.
- [x] Automatic version update system.

## Planned Features
- [ ] **Multi-Profile Support**: Allow users to create and manage multiple configuration profiles.
- [ ] **Multi-Provider Support**: Integrate with multiple AI providers (OpenAI, Anthropic/Claude, Google Gemini) in addition to Groq.
- [ ] **Profile Switching**: Add a command to quickly switch between profiles (e.g., `gs profile switch work`).
- [ ] **Custom Templates**: Support for user-defined commit message templates.
- [ ] **Pre-commit Hooks**: Integration with git hooks to automatically generate messages on `git commit`.
- [ ] **Interactive Message Editing**: Allow the user to tweak the generated message before finalizing.

## Improvements
- [ ] Enhance CLI UI/UX with more interactive components (similar to Charmbracelet/Bubble Tea).
- [ ] Improve error handling and system diagnostics.
- [ ] Add more comprehensive tests for different operating systems.
