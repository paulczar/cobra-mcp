# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
