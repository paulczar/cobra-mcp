# Cobra MCP Library

A pluggable library that enables any Cobra-based CLI application to expose its command structure as Model Context Protocol (MCP) tools and provide AI chat functionality.

## Features

- **Zero Configuration**: Automatically discover and expose all Cobra commands as MCP tools
- **Hierarchical Tool Structure**: Group commands by action (create, list, describe, etc.) to reduce tool count
- **Intelligent System Messages**: Auto-generate detailed system messages explaining tool usage patterns
- **Rich Tool Schemas**: Include flag descriptions and enum values directly in tool schemas for better AI accuracy
- **Dangerous Command Safety**: Configure dangerous commands that require explicit confirmation before execution
- **Debug Mode**: Chat client includes debug mode to show tool calls and parameters
- **Pluggable Architecture**: Easy integration into any existing Cobra CLI with minimal code changes
- **Dual Modes**: Support both MCP server (for MCP clients) and chat client (for direct AI interaction)
- **In-Process Execution**: Execute Cobra commands directly in-process for optimal performance

## Installation

```bash
go get github.com/paulczar/cobra-mcp/pkg
```

## Quick Start

### Basic Integration

```go
package main

import (
	"github.com/spf13/cobra"
	cobra_mcp "github.com/paulczar/cobra-mcp/pkg"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mycli",
		Short: "My CLI tool",
	}

	// Register your commands...
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)

	// Add MCP commands (these are added directly to root, not under a subcommand)
	rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, &cobra_mcp.ServerConfig{
		Name:       "mycli-mcp-server",
		ToolPrefix: "mycli",
	}))

	rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, &cobra_mcp.ChatConfig{
		Model: "gpt-4",
	}))

	rootCmd.Execute()
}
```

## Usage

### MCP Server

Start the MCP server over stdin:

```bash
mycli mcp start
```

Start the MCP server over HTTP:

```bash
mycli mcp stream --port 8080
```

Export available MCP tools as JSON:

```bash
mycli mcp tools
```

### Chat Client

Start an interactive chat session:

```bash
mycli chat --api-key YOUR_API_KEY
```

Or use environment variable:

```bash
export OPENAI_API_KEY=your_key
mycli chat
```

Process a single message:

```bash
mycli chat --message "List all clusters"
```

Enable debug mode to see tool calls and parameters:

```bash
mycli chat --debug --message "Create a cluster"
```

Read from stdin:

```bash
echo "List clusters" | mycli chat --stdin
```

Print the system message:

```bash
mycli chat system-message
```

## Configuration

### Server Configuration

```go
config := &cobra_mcp.ServerConfig{
	Name:          "mycli-mcp-server",
	Version:       "1.0.0",
	ToolPrefix:    "mycli",
	EnableResources: true,
	CustomActions: []string{"create", "list", "describe", "delete"},
	StandaloneCmds: []string{"version", "help"},
	// Dangerous commands that require explicit confirmation
	DangerousCommands: []string{"delete", "destroy"},
}
```

### Chat Configuration

```go
config := &cobra_mcp.ChatConfig{
	APIKey:            "your-api-key",
	APIURL:            "", // Optional custom API URL
	Model:             "gpt-4",
	Debug:             false, // Enable debug output showing tool calls and parameters
	SystemMessage:     "", // Optional custom system message
	SystemMessageFile: "", // Optional file path for system message
}
```

## API Reference

### NewMCPCommand

Creates a new Cobra command group with MCP subcommands (`mcp start`, `mcp stream`, `mcp tools`).

```go
func NewMCPCommand(rootCmd *cobra.Command, config *ServerConfig) *cobra.Command
```

The command provides three subcommands:
- `mcp start` - Start MCP server over stdin/stdout
- `mcp stream` - Start MCP server over HTTP (with `--port` flag, default: 8080)
- `mcp tools` - Export available MCP tools as JSON

### NewMCPServeCommand (Deprecated)

Creates a new Cobra command for serving MCP over stdio or HTTP. **Deprecated**: Use `NewMCPCommand` instead.

```go
func NewMCPServeCommand(rootCmd *cobra.Command, config *ServerConfig) *cobra.Command
```

### NewChatCommand

Creates a new Cobra command for AI chat with tool calling.

```go
func NewChatCommand(rootCmd *cobra.Command, config *ChatConfig) *cobra.Command
```

### GenerateSystemMessage

Generates a system message for the chat client.

```go
func GenerateSystemMessage(config *SystemMessageConfig) string
```

### NewServer

Creates a new MCP server instance.

```go
func NewServer(rootCmd *cobra.Command, config *ServerConfig) *Server
```

### NewChatClient

Creates a new chat client instance.

```go
func NewChatClient(server *Server, config *ChatConfig) (*ChatClient, error)
```

## Best Practices

### Error Handling - Avoid `os.Exit()`

**⚠️ CRITICAL**: Commands executed through MCP/chat run **in-process**. If your commands call `os.Exit()`, it will **terminate the entire MCP server or chat client process**, preventing further interaction.

**Use `RunE:` instead of `Run:`** to return errors instead of calling `os.Exit()`:

```go
// ❌ BAD - os.Exit() terminates the MCP/chat process
var listCmd = &cobra.Command{
    Use: "list",
    Run: func(cmd *cobra.Command, args []string) {
        if err := doSomething(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)  // ❌ This kills the MCP server!
        }
    },
}

// ✅ GOOD - Returns error instead
var listCmd = &cobra.Command{
    Use: "list",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := doSomething(); err != nil {
            return fmt.Errorf("error: %w", err)  // ✅ Error is handled gracefully
        }
        return nil
    },
}
```

**Why this matters:**
- Commands execute in the same process as the MCP server/chat client
- `os.Exit()` immediately terminates the entire process (no cleanup, no return)
- The executor cannot capture output or continue the session after `os.Exit()`
- Using `RunE:` allows proper error handling and the session continues

### Command Output

When writing command output in your Cobra commands, **prefer using `cmd.Println()` or `cmd.Printf()`** instead of `fmt.Println()` or direct writes to `os.Stdout`:

```go
// ✅ Preferred - respects output redirection
cmd.Println(`{"id": "cluster-123"}`)

// ✅ Also works - captured automatically
fmt.Println(`{"id": "cluster-123"}`)
```

The library automatically captures both Cobra's output methods (`cmd.Println`, `cmd.Printf`, etc.) and direct writes (`fmt.Println`, `os.Stdout.Write`, etc.) to ensure compatibility. However, using `cmd.Println` is recommended as it:
- Respects output redirection (important for MCP protocol)
- Follows Cobra best practices
- Works correctly with all Cobra features

### Flag Descriptions

The library automatically extracts flag descriptions and includes them in tool schemas. To improve AI accuracy:

- **Provide detailed flag descriptions**: Include enum values in descriptions (e.g., `"Cluster size: Small, Medium, or Large (required)"`)
- **Use clear descriptions**: The AI uses flag descriptions to understand what each flag does
- **Mark required flags**: Use `cmd.MarkFlagRequired()` so the AI knows which flags are mandatory

Example:
```go
createCmd.Flags().String("size", "", "Cluster size: Small, Medium, or Large (required)")
createCmd.MarkFlagRequired("size")
```

The library will automatically:
- Extract enum values ("Small", "Medium", "Large") from the description
- Include the flag description in the tool schema
- Mark the flag as required in the schema

## Examples

See the `examples/` directory for complete examples:
- `examples/basic/` - Basic integration example
- `examples/advanced/` - Advanced customization example

## Implementation Guide

For detailed step-by-step instructions on integrating MCP and Chat functionality into your existing Cobra CLI, see **[IMPLEMENTATION.md](IMPLEMENTATION.md)**.

This guide is designed for AI coding agents and developers, providing:
- Step-by-step integration instructions
- Command structure analysis
- Configuration examples
- Common patterns and troubleshooting
- Complete code examples

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes and version history.

## License

See LICENSE file for details.
