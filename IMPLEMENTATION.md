# Implementation Guide for AI Coding Agents

Use this guide to integrate MCP and Chat functionality into an existing Cobra-based CLI application.

## Step 1: Add Dependency

Add the library to your `go.mod`:

```bash
go get github.com/paulczar/cobra-mcp/pkg
```

## Step 2: Import the Library

In your main application file (typically `main.go` or `cmd/root.go`), add the import:

```go
import (
    // ... existing imports ...
    cobra_mcp "github.com/paulczar/cobra-mcp/pkg"
)
```

## Step 3: Identify Your Command Structure

Analyze your existing Cobra command structure:

1. **Identify actions**: Look for command patterns like `create`, `list`, `describe`, `delete`, `edit`, `upgrade`, etc.
   - Example: `create cluster`, `create machinepool` → action: `create`
   - Example: `list clusters`, `list nodes` → action: `list`

2. **Identify resources**: These are the subcommands under actions.
   - Example: Under `create`, you might have `cluster`, `machinepool`, `user`
   - Example: Under `list`, you might have `clusters`, `nodes`, `users`

3. **Identify standalone commands**: Commands that don't follow the action/resource pattern.
   - Example: `version`, `whoami`, `login`, `logout`

## Step 4: Configure Server Settings

Create a `ServerConfig` with appropriate settings. **Most fields are optional** - the library provides sensible defaults:

```go
serverConfig := &cobra_mcp.ServerConfig{
    Name:            "yourcli-mcp-server",  // MCP server name (optional, defaults to "{cli-name}-mcp-server")
    Version:         "1.0.0",               // Server version (optional, defaults to "1.0.0")
    ToolPrefix:      "yourcli",             // Tool name prefix (optional, defaults to CLI name)
    EnableResources: true,                   // Enable MCP resources (optional, defaults to true)
    // CustomActions and StandaloneCmds are OPTIONAL - see below
}
```

**Key decisions:**
- `ToolPrefix`: Should match your CLI name (e.g., "kubectl", "terraform", "rosa"). Optional - defaults to your CLI name.
- `CustomActions`: **Optional** - Leave empty/nil for auto-detection. The library will automatically detect any first-level command as an action. Only specify if you want to restrict which commands are treated as actions (whitelist mode).
- `StandaloneCmds`: **Optional** - Leave empty/nil for auto-detection. The library will automatically detect standalone commands (commands with no subcommands). Only specify if you want to explicitly mark certain commands as standalone.

**Minimal configuration example** (works for most CLIs):
```go
serverConfig := &cobra_mcp.ServerConfig{
    ToolPrefix: "yourcli",  // Only specify if you want a different prefix
    // CustomActions and StandaloneCmds are auto-detected - no need to specify!
    // ExecutionMode defaults to "in-process" - see below for options
}
```

**Execution Mode Configuration**:

The library supports three execution modes to handle commands that may call `os.Exit()`:

```go
serverConfig := &cobra_mcp.ServerConfig{
    ToolPrefix:    "yourcli",
    ExecutionMode: "auto", // Options: "in-process", "sub-process", or "auto"
}
```

- **`"in-process"`** (default): Execute all commands directly in-process for optimal performance. Fast but vulnerable to `os.Exit()` calls that terminate the MCP/chat process.
- **`"sub-process"`**: Execute all commands in a sub-process. Safer (can't kill parent process) but slower due to process spawn overhead.
- **`"auto"`** (recommended): Auto-detect commands using `Run:` (no `RunE:`) and execute them in sub-process, while using in-process for commands with `RunE:`. Best of both worlds - fast for safe commands, safe for risky ones.

**Recommended**: Use `"auto"` mode if your CLI has commands that use `Run:` with `os.Exit()`:

```go
serverConfig := &cobra_mcp.ServerConfig{
    ToolPrefix:    "yourcli",
    ExecutionMode: "auto", // Automatically protect against os.Exit()
}
```

**Note**: The executable path for sub-process execution is automatically detected using `os.Executable()` (current running binary). No additional configuration needed.

**Auto-detection behavior:**
- **Actions**: Any first-level command is automatically detected as an action (e.g., `create`, `list`, `deploy`, `status` - all work automatically)
- **Standalone commands**: Commands with no subcommands are automatically detected as standalone (e.g., `version`, `help`, `status`)

**When you need CustomActions (whitelist mode):**
- You want to restrict which commands are exposed as actions (e.g., exclude certain commands from being actions)
- Example: Only expose `create`, `list`, `delete` but not `internal-debug`

**When you need StandaloneCmds (explicit mode):**
- You want to explicitly mark certain commands as standalone, even if they have subcommands
- Rarely needed since auto-detection handles this correctly

## Step 5: Add MCP Command

Add the MCP command group to your root command:

```go
rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, serverConfig))
```

This adds an `mcp` command group with three subcommands:
- `mcp start` - Start MCP server over stdin/stdout
- `mcp stream` - Start MCP server over HTTP (with `--port` flag, default: 8080)
- `mcp tools` - Export available MCP tools as JSON

## Step 6: Add Chat Command (Optional)

If you want AI chat functionality, add the chat command:

```go
chatConfig := &cobra_mcp.ChatConfig{
    Model: "gpt-4",  // or "gpt-4-turbo", "gpt-3.5-turbo", etc.
    Debug: false,    // Set to true for debug logging
}

rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, chatConfig))
```

## Step 7: Handle Commands with `os.Exit()`

**⚠️ CRITICAL**: In default `"in-process"` mode, commands executed through MCP/chat run **in-process**. If your commands call `os.Exit()`, it will **terminate the entire MCP server or chat client process**.

**Note**: If you're using `ExecutionMode: "auto"` or `"sub-process"`, you don't need to worry about this - those modes automatically protect against `os.Exit()` calls. The library will also suppress warnings about `Run:` commands when using these modes.

**Option 1: Use `ExecutionMode: "auto"` (Recommended)**

The easiest solution is to use auto mode, which automatically executes commands with `Run:` in sub-process:

```go
serverConfig := &cobra_mcp.ServerConfig{
    ExecutionMode: "auto", // Automatically protects against os.Exit()
}
```

**Option 2: Use `ExecutionMode: "sub-process"`**

Execute all commands in sub-process for maximum safety:

```go
serverConfig := &cobra_mcp.ServerConfig{
    ExecutionMode: "sub-process", // All commands in sub-process
}
```

**Option 3: Migrate to `RunE:` (Best Practice)**

Always use `RunE:` to return errors instead of calling `os.Exit()`:

```go
// ❌ BAD - os.Exit() terminates the MCP/chat process
var listCmd = &cobra.Command{
    Use: "list",
    Run: func(cmd *cobra.Command, args []string) {
        if err := listClusters(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)  // ❌ This kills the MCP server!
        }
    },
}

// ✅ GOOD - Returns error instead
var listCmd = &cobra.Command{
    Use: "list",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := listClusters(); err != nil {
            return fmt.Errorf("failed to list clusters: %w", err)  // ✅ Error handled gracefully
        }
        return nil
    },
}
```

**Why this matters:**
- `os.Exit()` immediately terminates the entire process (no cleanup, no return)
- The executor cannot capture output or continue the session after `os.Exit()`
- Using `RunE:` allows proper error handling and the session continues

## Step 8: Verify Command Output Methods

Ensure your commands use appropriate output methods:

**Preferred (recommended):**
```go
cmd.Println(`{"id": "123", "name": "example"}`)
cmd.Printf("Status: %s\n", status)
```

**Also supported (but less ideal):**
```go
fmt.Println(`{"id": "123"}`)
os.Stdout.Write([]byte("output"))
```

The library captures both, but `cmd.Println()` is preferred for proper output redirection.

## Step 9: Test the Integration

1. **Test MCP Server over stdin:**
   ```bash
   yourcli mcp start
   ```

2. **Test MCP Server over HTTP:**
   ```bash
   yourcli mcp stream --port 8080
   ```

3. **Export MCP tools as JSON:**
   ```bash
   yourcli mcp tools
   ```

4. **Test Chat (if added):**
   ```bash
   export OPENAI_API_KEY=your_key
   yourcli chat
   ```

5. **View system message:**
   ```bash
   yourcli chat system-message
   ```

## Step 10: Customize (Optional)

### Custom System Message

If you want to customize the AI's instructions:

```go
chatConfig := &cobra_mcp.ChatConfig{
    Model: "gpt-4",
    SystemMessageFile: "path/to/custom-system-message.txt",
}
```

### Custom Actions (Whitelist Mode)

**By default, all first-level commands are auto-detected as actions.** You only need to specify `CustomActions` if you want to restrict which commands are exposed (whitelist mode).

If you want to restrict which commands are treated as actions:

```go
serverConfig := &cobra_mcp.ServerConfig{
    ToolPrefix:    "terraform",
    CustomActions: []string{"apply", "plan", "destroy", "refresh"}, // Only these actions will be exposed
}
```

**Note:** If you specify `CustomActions`, only those actions will be recognized (whitelist mode). All other commands will be ignored. This is useful if you want to exclude certain commands from being exposed as MCP tools.

## Step 11: Verify Tool Generation

After integration, verify tools are generated correctly:

1. Export available tools: `yourcli mcp tools` (shows all tools as JSON)
2. Or start the MCP server: `yourcli mcp start`
3. Send an `initialize` request
4. Send a `tools/list` request
5. Verify your commands appear as hierarchical tools (e.g., `yourcli_list`, `yourcli_create`)

## Common Patterns

### Pattern 1: Minimal Configuration (Recommended)

For most CLIs, you only need to specify `ToolPrefix` (or nothing at all):

```go
rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, &cobra_mcp.ServerConfig{
    ToolPrefix: "mycli",  // Optional - defaults to CLI name
}))
```

Or even simpler (uses all defaults):
```go
rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, &cobra_mcp.ServerConfig{}))
```

### Pattern 2: CLI with Custom Actions

```go
rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, &cobra_mcp.ServerConfig{
    Name:          "mycli-mcp-server",
    ToolPrefix:    "mycli",
    CustomActions: []string{"deploy", "rollback", "scale"},
}))
```

### Pattern 3: Full Integration with Chat

```go
// MCP Server
rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, &cobra_mcp.ServerConfig{
    Name:       "mycli-mcp-server",
    ToolPrefix: "mycli",
}))

// Chat Client
rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, &cobra_mcp.ChatConfig{
    Model: "gpt-4",
}))
```

## Troubleshooting

1. **Tools not appearing**:
   - Check that your commands follow the action/resource pattern (e.g., `create cluster`, `list nodes`)
   - If you specified `CustomActions`, make sure your action is in the whitelist (auto-detection is disabled when `CustomActions` is set)
   - Standalone commands (like `version`, `help`) are auto-detected if they have no subcommands
   - Use `yourcli mcp tools` to see all available tools as JSON
2. **MCP server or chat client terminates unexpectedly**:
   - **⚠️ CRITICAL**: Commands using `Run:` with `os.Exit()` will terminate the entire MCP/chat process in default `"in-process"` mode
   - **Note**: If you see warnings about commands using `Run:`, it means you're in `"in-process"` mode. The warnings are automatically suppressed when using `"auto"` or `"sub-process"` modes.
   - **Solution 1** (Recommended): Use `"auto"` execution mode to automatically execute commands with `Run:` in sub-process:
     ```go
     serverConfig := &cobra_mcp.ServerConfig{
         ExecutionMode: "auto", // Auto-protect against os.Exit(), no warnings shown
     }
     ```
   - **Solution 2**: Use `"sub-process"` execution mode to execute all commands in sub-process:
     ```go
     serverConfig := &cobra_mcp.ServerConfig{
         ExecutionMode: "sub-process", // All commands in sub-process, no warnings shown
     }
     ```
   - **Solution 3**: Migrate commands to use `RunE:` instead of `Run:` and return errors instead of calling `os.Exit()`
   - See Step 7 above for examples
3. **Output not captured**: Ensure commands use `cmd.Println()` or `fmt.Println()` (both are supported)
4. **MCP server hangs**: Verify commands don't write directly to `os.Stderr` (use `cmd.PrintErr()` instead)
5. **Chat not working**: Verify `OPENAI_API_KEY` is set or passed via `--api-key` flag
6. **Custom actions not working**: If you specified `CustomActions`, remember it acts as a whitelist - only actions in the list will be recognized. Auto-detection is disabled when `CustomActions` is set.

## Complete Example

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

    // Your existing commands
    rootCmd.AddCommand(createCmd)
    rootCmd.AddCommand(listCmd)
    rootCmd.AddCommand(describeCmd)
    // ... etc

    // Add MCP support (minimal configuration - uses defaults for actions)
    rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, &cobra_mcp.ServerConfig{
        ToolPrefix: "mycli",  // Optional - defaults to CLI name
    }))

    // Add Chat support (optional)
    rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, &cobra_mcp.ChatConfig{
        Model: "gpt-4",
    }))

    rootCmd.Execute()
}
```
