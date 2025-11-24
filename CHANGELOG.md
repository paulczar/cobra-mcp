# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] - 2025-01-XX

### Added
- `SystemMessageAppend` field in `ChatConfig` to append custom content to the generated system message
- Allows CLI applications to add additional instructions (e.g., output limitation guidelines) without overriding the entire system message
- Appended content is automatically included when viewing system message via `chat system-message` command

## [1.3.0] - 2025-01-XX

### Added
- Sub-process execution mode to handle CLIs that use `Run:` with `os.Exit()`
- `ExecutionMode` configuration option in `ServerConfig` with three modes:
  - `"in-process"` (default): Execute all commands in-process for optimal performance
  - `"sub-process"`: Execute all commands in sub-process to protect against `os.Exit()`
  - `"auto"`: Auto-detect commands using `Run:` and execute them in sub-process, others in-process
- `NewCommandExecutorWithMode()` function to create executor with specific execution mode
- `ExecuteSubProcess()` method for sub-process command execution
- Automatic executable path detection using `os.Executable()` (no configuration needed)
- Conditional warning system that only warns in `"in-process"` mode

### Changed
- Warnings about commands using `Run:` are now conditional - only shown in `"in-process"` mode
- Warning message updated to suggest using `ExecutionMode: "auto"` or `"sub-process"` as alternatives
- `CommandExecutor` now supports execution mode configuration
- `Execute()` method now routes to sub-process or in-process execution based on mode

### Fixed
- Commands using `Run:` with `os.Exit()` no longer terminate MCP/chat process when using `"auto"` or `"sub-process"` modes
- Parent process survives `os.Exit()` calls when commands execute in sub-process

## [1.2.0] - 2025-01-XX

### Added
- New `NewMCPCommand` function that creates an `mcp` command group with subcommands:
  - `mcp start` - Start MCP server over stdin/stdout
  - `mcp stream` - Start MCP server over HTTP (with `--port` flag)
  - `mcp tools` - Export available MCP tools as JSON
- `mcp tools` command for exporting tool definitions as JSON for debugging and integration

### Changed
- `NewMCPServeCommand` is now deprecated (kept for backward compatibility)
- Updated all documentation to reflect new command structure
- Examples updated to use `NewMCPCommand`

## [1.1.0] - 2025-01-XX

### Added
- Detection and warnings for commands using `Run:` instead of `RunE:`
- Automatic warnings when MCP server or chat client starts if commands may call `os.Exit()`
- E2E tests for `Run:` detection and warning system
- Excludes built-in Cobra `help` command from warnings (users can't control it)

### Fixed
- Tool schema description now correctly marks `flags` parameter as optional instead of "REQUIRED for most commands"
- This prevents AI models from being confused when flags are not provided
- Chat debug mode now displays the model being used for API calls
- Chat command now respects ChatConfig.Model when --model flag is not explicitly provided
- Previously, the default value of the --model flag ("gpt-4") would override ChatConfig.Model even when the flag wasn't set

### Changed
- Updated examples to use `RunE:` instead of `Run:` to follow best practices
- Documentation updated with warnings about avoiding `os.Exit()` in commands

## [1.0.0] - 2025-11-19

### Added
- Initial release of cobra-mcp library
- Automatic command discovery and MCP tool generation
- Hierarchical tool structure (action-based grouping)
- Auto-detection of actions and standalone commands
- In-process command execution for optimal performance
- Output capture for both Cobra methods and direct stdout writes
- Rich tool schemas with flag descriptions and enum values
- Dangerous command safety with explicit confirmation requirements
- Debug mode for chat client showing tool calls and parameters
- Enhanced help tool with structured, AI-friendly output
- System message generation with CLI help text integration
- Support for both MCP server (stdio/HTTP) and chat client modes
- Comprehensive e2e test suite
- Makefile for development tasks
- Complete documentation (README, DESIGN, IMPLEMENTATION)

### Features
- Zero-configuration auto-detection of CLI structure
- Pluggable architecture for easy integration
- Support for custom actions and standalone commands (whitelist mode)
- Flag description extraction and enum value parsing
- JSON Schema type conversion for tool parameters
- Parameter discovery guidance for AI agents
- Rate limit error handling with helpful messages
