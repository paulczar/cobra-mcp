# Design Document: Pluggable MCP Library for Cobra-Based CLIs

## 1. Overview

This document describes the design for a reusable, pluggable library that enables any Cobra-based CLI application to expose its command structure as Model Context Protocol (MCP) tools and provide AI chat functionality. The library automatically discovers commands, generates hierarchical tool definitions, creates comprehensive system messages, and provides both MCP server and chat client capabilities.

### 1.1 Goals

- **Zero Configuration**: Automatically discover and expose all Cobra commands as MCP tools
- **Hierarchical Tool Structure**: Group commands by action (create, list, describe, etc.) to reduce tool count from hundreds to dozens
- **Intelligent System Messages**: Auto-generate detailed system messages explaining tool usage patterns
- **Pluggable Architecture**: Easy integration into any existing Cobra CLI with minimal code changes
- **Dual Modes**: Support both MCP server (for MCP clients) and chat client (for direct AI interaction)
- **In-Process Execution**: Execute Cobra commands directly in-process by calling their Run functions, avoiding subprocess overhead

### 1.2 Key Features

- Automatic command discovery from Cobra command tree
- Hierarchical tool organization (action-based grouping)
- Dynamic system message generation with tool-specific instructions
- Rich tool schemas with flag descriptions and enum values extracted from flag descriptions
- Dangerous command safety with explicit confirmation requirements
- Debug mode for chat client to show tool calls and parameters
- Support for both stdio and HTTP transports for MCP server
- OpenAI-compatible chat client with tool calling
- Resource registry for exposing CLI data as MCP resources
- Flag and argument extraction from Cobra commands
- JSON output handling for structured responses

## 2. Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Cobra CLI Application                     │
│                  (Root Command + Subcommands)                │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              Cobra MCP Library (This Library)               │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Command    │  │     Tool     │  │  Resource    │     │
│  │  Executor    │  │   Registry   │  │  Registry   │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                 │                  │              │
│         └─────────────────┼──────────────────┘              │
│                           │                                  │
│  ┌───────────────────────┴───────────────────────┐          │
│  │              MCP Server                       │          │
│  │  (StdioTransport / HTTPTransport)             │          │
│  └───────────────────────┬───────────────────────┘          │
│                           │                                  │
│  ┌───────────────────────┴───────────────────────┐          │
│  │            Chat Client                        │          │
│  │  (OpenAI API Integration)                     │          │
│  └───────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              MCP Clients / AI Assistants                    │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Component Responsibilities

1. **CommandExecutor**: Discovers commands, executes them directly in-process by calling Cobra command Run functions, extracts flags/args
2. **ToolRegistry**: Converts Cobra commands to hierarchical MCP tools
3. **ResourceRegistry**: Exposes CLI data as MCP resources (optional)
4. **Server**: Wraps MCP server implementation, registers tools/resources
5. **ChatClient**: Provides OpenAI-compatible chat interface with tool calling

## 3. Core Components

### 3.1 CommandExecutor

**Purpose**: Discover and execute Cobra commands either in-process or in sub-process, depending on execution mode configuration.

**Execution Modes**:
- **`"in-process"`** (default): Execute commands directly in-process for optimal performance. Fast but vulnerable to `os.Exit()` calls.
- **`"sub-process"`**: Execute all commands in a sub-process. Safer (can't kill parent process) but slower due to process spawn overhead.
- **`"auto"`**: Auto-detect commands using `Run:` (no `RunE:`) and execute them in sub-process, while using in-process for commands with `RunE:`. Best of both worlds.

**⚠️ Important Limitation**: In default `"in-process"` mode, if a command calls `os.Exit()`, it will **terminate the entire MCP server or chat client process**. Use `"auto"` or `"sub-process"` execution mode, or migrate commands to use `RunE:` (return errors) instead of `Run:` with `os.Exit()` calls.

**Key Responsibilities**:
- Traverse Cobra command tree to discover all commands
- Extract flag information (name, type, description, required)
- Execute commands either in-process or in sub-process based on execution mode
- Route execution based on command type (`Run:` vs `RunE:`) in auto mode
- Capture command output by redirecting stdout/stderr to buffers (in-process) or via pipes (sub-process)
- Automatically add JSON output flags when supported

**API**:
```go
type CommandExecutor struct {
    rootCmd       *cobra.Command
    executionMode string // "in-process", "sub-process", or "auto"
}

func NewCommandExecutor(rootCmd *cobra.Command) *CommandExecutor // Defaults to "in-process"
func NewCommandExecutorWithMode(rootCmd *cobra.Command, executionMode string) *CommandExecutor

type CommandInfo struct {
    Path        []string  // Command path: ["create", "cluster"]
    Description string    // Short description
    Use         string    // Usage string
    Long        string    // Long description
    Flags       []FlagInfo
}

type FlagInfo struct {
    Name        string
    Shorthand   string
    Description string
    Type        string  // "string", "bool", "int", etc.
    Required    bool
}

type ExecuteResult struct {
    Stdout   string
    Stderr   string
    ExitCode int
    Error    error
}

func (e *CommandExecutor) GetAllCommands() []CommandInfo
func (e *CommandExecutor) Execute(commandPath []string, flags map[string]interface{}) (*ExecuteResult, error)
func (e *CommandExecutor) FindCommand(path []string) (*cobra.Command, []string, error)
func (e *CommandExecutor) ExecuteSubProcess(commandPath []string, flags map[string]interface{}) (*ExecuteResult, error)
func (e *CommandExecutor) shouldUseSubProcess(cmd *cobra.Command) bool // Internal helper
```

**Implementation Details**:
- Use `cobra.Command.Find()` to locate commands in the tree
- Extract flags from both `Flags()` and `PersistentFlags()`
- Detect required flags using `cobra.BashCompOneRequiredFlag` annotation
- Execute commands directly in-process by calling `rootCmd.ExecuteContext(ctx)` with the full command path
- Capture output by redirecting `rootCmd.SetOut()` and `rootCmd.SetErr()` to `bytes.Buffer`
- **Capture direct stdout writes**: Redirect `os.Stdout` using a pipe to capture `fmt.Println()`, `os.Stdout.Write()`, etc. that bypass Cobra's output redirection
- Redirect `os.Stderr` to `/dev/null` to prevent any direct writes from breaking JSON-RPC protocol
- Merge captured direct stdout writes with Cobra's output buffers
- Restore original stdout/stderr after execution to avoid interference
- Set `SURVEY_FORCE_NO_INTERACTIVE=1` to disable interactive prompts
- Set flags programmatically on the target command using `flag.Value.Set()` and `cmd.Flags().Set()`
- Automatically add `-o json` or `--output=json` if command supports output flag
- Execute from root command with full path: `rootCmd.SetArgs(["list", "clusters"])` then `rootCmd.ExecuteContext(ctx)`

**Output Capture Strategy**:
- **Cobra methods** (`cmd.Println()`, `cmd.Printf()`, etc.): Captured via `rootCmd.SetOut()` redirection
- **Direct writes** (`fmt.Println()`, `os.Stdout.Write()`, etc.): Captured via `os.Stdout` pipe redirection
- Both are merged into a single output buffer before returning results
- **Best Practice**: Commands should use `cmd.Println()` as it respects output redirection and follows Cobra conventions, but direct writes are supported for compatibility

**Sub-Process Execution**:
- When execution mode is `"sub-process"` or `"auto"` (and command uses `Run:`), commands are executed in a separate process
- Uses `os.Executable()` to get the current running binary path automatically
- Builds command line arguments from command path, flags, and positional args
- Captures stdout/stderr via pipes using `exec.CommandContext()`
- Captures exit code from `cmd.Wait()`
- Handles context cancellation/timeout
- Preserves same `ExecuteResult` interface for compatibility
- **Benefits**: Commands calling `os.Exit()` cannot kill the parent MCP/chat process
- **Trade-offs**: Process spawn overhead, but safer for commands that may call `os.Exit()`

**⚠️ Critical: Avoid `os.Exit()` in Commands**:
- In default `"in-process"` mode, commands execute **in-process** - `os.Exit()` terminates the entire MCP/chat process
- Use `"auto"` execution mode to automatically execute commands with `Run:` in sub-process (recommended)
- Use `"sub-process"` execution mode to execute all commands in sub-process (safest)
- Use `RunE:` instead of `Run:` to return errors instead of calling `os.Exit()` (best practice)
- The executor cannot capture output or continue the session after `os.Exit()` is called in in-process mode
- Sub-process execution isolates `os.Exit()` calls and prevents parent process termination

**Warning System**:
- The library automatically detects commands using `Run:` (no `RunE:`) and warns about potential `os.Exit()` issues
- **Warnings are only shown** when `ExecutionMode` is `"in-process"` (default)
- When `ExecutionMode` is `"auto"` or `"sub-process"`, warnings are suppressed because these modes protect against `os.Exit()`
- The warning message suggests using `ExecutionMode: "auto"` or `"sub-process"` as an alternative to migrating commands

### 3.2 ToolRegistry

**Purpose**: Convert discovered Cobra commands into hierarchical MCP tool definitions.

**Key Responsibilities**:
- Discover all commands via CommandExecutor
- Group commands by action (create, list, describe, delete, etc.)
- Generate hierarchical tools with `resource` parameter
- Create standalone tools for commands without subcommands
- Generate detailed tool descriptions with examples
- Map Cobra flag types to JSON Schema types

**API**:
```go
type ToolRegistry struct {
    executor *CommandExecutor
    tools    map[string]*ToolDefinition
}

type ToolDefinition struct {
    Name        string
    Description string
    Parameters  map[string]ParameterInfo
}

type ParameterInfo struct {
    Type        string
    Description string
    Required    bool
}

func NewToolRegistry(rootCmd *cobra.Command) *ToolRegistry
func (tr *ToolRegistry) GetHierarchicalTools() []map[string]interface{}
func (tr *ToolRegistry) CallTool(toolName string, arguments map[string]interface{}) (map[string]interface{}, error)
```

**Hierarchical Tool Structure**:

Instead of exposing every command as a separate tool (e.g., `rosa_create_cluster`, `rosa_create_machinepool`, etc.), the library groups commands by action:

- **Hierarchical Tools**: `{prefix}_create`, `{prefix}_list`, `{prefix}_describe`, etc.
  - Require `resource` parameter (enum of available resources)
  - Accept `flags` object for command flags
  - Accept optional `args` array for positional arguments

- **Standalone Tools**: `{prefix}_whoami`, `{prefix}_version`, etc.
  - No `resource` parameter needed
  - Accept `flags` object

- **Special Tools**: `{prefix}_help`
  - Accepts `command` string or `resource` string

**Tool Naming Convention**:
- Tool names use format: `{prefix}_{action}` where prefix is configurable (default: CLI name)
- Example: `rosa_list`, `rosa_create`, `kubectl_get`, `terraform_apply`

**Tool Schema Generation**:

Each hierarchical tool has this schema:
```json
{
  "type": "object",
  "properties": {
    "resource": {
      "type": "string",
      "description": "REQUIRED: The resource type to {action}. Must be one of: {resources}",
      "enum": ["resource1", "resource2", ...]
    },
    "flags": {
      "type": "object",
      "description": "Optional: Command flags as key-value pairs (flag names without '--' prefix). Provide an empty object {} if no flags are needed.",
      "properties": {
        "flagName": {
          "type": "string",
          "description": "Flag description extracted from Cobra command",
          "enum": ["value1", "value2"]  // Extracted from description if present
        }
      },
      "required": ["flagName"],  // Based on cmd.MarkFlagRequired()
      "additionalProperties": true
    },
    "args": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Optional positional arguments (rarely used)"
    }
  },
  "required": ["resource"]
}
```

**Flag Schema Enhancement**:
- Flag descriptions are extracted from Cobra command flag definitions
- Enum values are automatically parsed from descriptions (e.g., "Small, Medium, or Large" → ["Small", "Medium", "Large"])
- Required flags are marked based on `cmd.MarkFlagRequired()` calls
- Flag types are converted from Go/pflag types to JSON Schema types (bool → boolean, int64 → integer, etc.)

**Action Detection**:

The library automatically detects actions by analyzing command paths:
- Commands like `create cluster`, `create machinepool` → action: `create`
- Commands like `list clusters`, `list machinepools` → action: `list`
- Commands like `describe cluster` → action: `describe`

Actions are configurable but default to common patterns:
- `create`, `list`, `describe`, `delete`, `edit`, `upgrade`
- `grant`, `revoke`, `verify`
- `login`, `logout`, `whoami`, `version`, `help`

### 3.3 ResourceRegistry (Optional)

**Purpose**: Expose CLI data as MCP resources for read-only access.

**API**:
```go
type ResourceRegistry struct {
    executor *CommandExecutor
}

type ResourceDefinition struct {
    URI         string
    Name        string
    Description string
    MimeType    string
}

func NewResourceRegistry(executor *CommandExecutor) *ResourceRegistry
func (rr *ResourceRegistry) GetResources() []ResourceDefinition
func (rr *ResourceRegistry) ReadResource(uri string) (string, string, error)
```

**Implementation**:
- Resources use URI scheme: `{prefix}://{resource-type}[/{id}]`
- Example: `rosa://clusters`, `rosa://cluster/my-cluster-id`
- Resources are optional and can be customized per CLI

### 3.4 Server

**Purpose**: Wrap MCP server implementation and register tools/resources.

**API**:
```go
type Server struct {
    rootCmd          *cobra.Command
    toolRegistry     *ToolRegistry
    resourceRegistry *ResourceRegistry
    mcpServer        *mcp.Server
}

type ServerConfig struct {
    Name        string  // Server name (default: "{cli-name}-mcp-server")
    Version     string  // Server version (default: "1.0.0")
    ToolPrefix  string  // Tool name prefix (default: CLI name)
    EnableResources bool // Enable resource registry (default: true)
}

func NewServer(rootCmd *cobra.Command, config *ServerConfig) *Server
func ServeStdio(rootCmd *cobra.Command, config *ServerConfig) error
func ServeHTTP(rootCmd *cobra.Command, port int, config *ServerConfig) error
```

**Implementation**:
- Uses `github.com/modelcontextprotocol/go-sdk/mcp` for MCP protocol
- Registers all hierarchical tools from ToolRegistry
- Optionally registers resources from ResourceRegistry
- Supports both stdio and HTTP transports
- Converts tool schemas to JSON Schema format

### 3.5 ChatClient

**Purpose**: Provide OpenAI-compatible chat interface with tool calling.

**API**:
```go
type ChatClient struct {
    client       openai.Client
    model        shared.ChatModel
    toolRegistry *ToolRegistry
    messages     []openai.ChatCompletionMessageParamUnion
    debug        bool
}

type ChatConfig struct {
    APIKey            string
    APIURL            string  // Optional custom API URL
    Model             string  // Default: "gpt-4"
    Debug             bool
    SystemMessage     string  // Optional custom system message (overrides generated)
    SystemMessageFile string  // Optional file path for system message (overrides generated)
    SystemMessageAppend string // Optional content to append to generated system message
}

func NewChatClient(server *Server, config *ChatConfig) (*ChatClient, error)
func (cc *ChatClient) RunChatLoop() error  // Interactive REPL
func (cc *ChatClient) ProcessMessage(userInput string) error  // Single message
```

**System Message Generation**:

The library generates comprehensive system messages that include:

1. **CLI Overview**:
   - CLI name and description from root command help
   - Purpose and use cases

2. **Tool Usage Instructions**:
   - Explanation of hierarchical vs standalone tools
   - How to use `resource` parameter
   - How to use `flags` object
   - Flag naming conventions (without `--` prefix)

3. **Command Help Text**:
   - Concise help text for each tool/resource
   - Usage examples from command help
   - Available flags and their descriptions

4. **Common Patterns**:
   - Examples of common operations
   - Resource-specific patterns
   - Flag usage examples

5. **Output Format**:
   - JSON output handling
   - How to parse and present results

6. **Error Handling**:
   - When to use help tool
   - How to handle errors
   - Troubleshooting guidance

7. **Safety Requirements** (if applicable):
   - Destructive operation warnings
   - Confirmation requirements
   - Best practices

**System Message Template**:

```go
func GenerateSystemMessage(config *SystemMessageConfig) string {
    // Config includes:
    // - CLI name and description
    // - Tool prefix
    // - Available actions and resources
    // - Common patterns
    // - Safety requirements
    // - Custom instructions
}
```

**Example Generated System Message**:

```
You are a helpful assistant for managing {CLI_NAME} resources.

TOOL USAGE:
- {CLI_NAME} tools use a hierarchical structure. Hierarchical tools ({prefix}_list, {prefix}_describe, {prefix}_create, etc.) REQUIRE a 'resource' parameter.
- ALWAYS use the 'flags' parameter with flag names (without '--' prefix) for command options.
- For {resource-type}-related operations, ALWAYS include flags={'{resource-key}': '{resource-name-or-id}'}.
- Standalone tools ({prefix}_whoami, {prefix}_version, etc.) don't need a 'resource' parameter.

COMMON PATTERNS:
- List {resources}: {prefix}_list with resource='{resource-plural}'
- Describe {resource}: {prefix}_describe with resource='{resource-singular}' and flags={'{resource-key}': 'name'}
- {other common patterns}

OUTPUT FORMAT:
- All commands return JSON output automatically. Parse and present results clearly to the user.
- When listing resources, summarize key information (name, ID, status) in a readable format.

ERROR HANDLING:
- If unsure about command syntax or available options, use {prefix}_help with the command path.
- If a tool call fails, check the error message and suggest using {prefix}_help if needed.

{custom_safety_requirements}

BEST PRACTICES:
- {best_practices}
```

## 4. Integration Points

### 4.1 Library Initialization

The library requires minimal integration:

```go
import "github.com/paulczar/cobra-mcp/pkg"

func main() {
    rootCmd := &cobra.Command{
        Use:   "mycli",
        Short: "My CLI tool",
    }

    // Register all your commands...
    rootCmd.AddCommand(createCmd)
    rootCmd.AddCommand(listCmd)
    // etc.

    // Add MCP commands (these are added directly to root, not under a subcommand)
    rootCmd.AddCommand(cobra_mcp.NewMCPCommand(rootCmd, &cobra_mcp.ServerConfig{
        Name: "mycli-mcp-server",
        ToolPrefix: "mycli",
    }))

    rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, &cobra_mcp.ChatConfig{
        Model: "gpt-4",
    }))

    // Usage: mycli mcp start          (start MCP server over stdin)
    // Usage: mycli mcp stream         (start MCP server over HTTP)
    // Usage: mycli mcp tools          (export available MCP tools as JSON)
    // Usage: mycli chat --api-key YOUR_KEY
    // Usage: mycli chat system-message  (to print the system message)

    rootCmd.Execute()
}
```

### 4.2 Command Structure Requirements

The library works with any Cobra command structure, but works best with:

1. **Consistent Naming**: Commands follow patterns like `{action} {resource}`
2. **Flag Support**: Commands use Cobra flags (not just positional args)
3. **JSON Output**: Commands support `--output json` or `-o json` flag
4. **Descriptions**: Commands have `Short` and optionally `Long` descriptions

### 4.3 Customization Points

The library provides several customization options:

1. **Tool Prefix**: Customize tool name prefix (default: CLI name)
2. **Action Detection**: Customize which commands are considered actions
3. **Standalone Commands**: Specify which commands don't need resources
4. **System Message**: Customize system message generation
5. **Resource Registry**: Enable/disable or customize resources
6. **Flag Mapping**: Customize how flags are mapped to tool parameters

## 5. Implementation Details

### 5.1 Command Discovery Algorithm

```go
func (e *CommandExecutor) GetAllCommands() []CommandInfo {
    commands := []CommandInfo{}
    e.traverseCommands(e.rootCmd, []string{}, &commands)
    return commands
}

func (e *CommandExecutor) traverseCommands(cmd *cobra.Command, path []string, commands *[]CommandInfo) {
    // Skip hidden commands and root
    if cmd.Hidden || cmd == e.rootCmd {
        if cmd == e.rootCmd {
            // Still traverse children
            for _, subCmd := range cmd.Commands() {
                e.traverseCommands(subCmd, []string{}, commands)
            }
        }
        return
    }

    // Add current command
    currentPath := append(path, cmd.Name())
    info := CommandInfo{
        Path:        currentPath,
        Description: cmd.Short,
        Use:         cmd.Use,
        Long:        cmd.Long,
        Flags:       e.extractFlags(cmd),
    }
    *commands = append(*commands, info)

    // Traverse subcommands
    for _, subCmd := range cmd.Commands() {
        e.traverseCommands(subCmd, currentPath, commands)
    }
}
```

### 5.2 Hierarchical Tool Generation

```go
func (tr *ToolRegistry) GetHierarchicalTools() []map[string]interface{} {
    tools := []map[string]interface{}{}

    // Detect actions from command paths
    actions := tr.detectActions()

    for _, action := range actions {
        resources := tr.getAvailableResources(action)

        if len(resources) == 0 {
            // Standalone command
            if tr.isStandaloneCommand(action) {
                tool := tr.createStandaloneTool(action)
                if tool != nil {
                    tools = append(tools, tool)
                }
            }
            continue
        }

        // Hierarchical tool
        tool := tr.createHierarchicalTool(action, resources)
        tools = append(tools, tool)
    }

    return tools
}
```

### 5.3 Tool Execution Flow

```
1. MCP client calls tool: {prefix}_list with {resource: "clusters", flags: {}}
2. ToolRegistry.CallTool() receives call
3. Parse tool name to extract action: "list"
4. Extract resource from arguments: "clusters"
5. Build command path: ["list", "clusters"]
6. Extract flags from arguments["flags"]
7. CommandExecutor.Execute(["list", "clusters"], flags)
8. CommandExecutor finds command in tree using rootCmd.Find()
9. CommandExecutor redirects rootCmd.SetOut() and rootCmd.SetErr() to buffers
10. CommandExecutor redirects os.Stderr to /dev/null to prevent protocol interference
11. CommandExecutor sets flags on target command
12. CommandExecutor sets args on root command: rootCmd.SetArgs(["list", "clusters"])
13. CommandExecutor calls rootCmd.ExecuteContext(ctx) directly (in-process)
14. Cobra internally calls the command's Run function
15. Command output is captured in buffers
16. Original stdout/stderr are restored
17. Return result as MCP format
```

### 5.4 In-Process Command Execution

The library executes Cobra commands directly in-process by calling their `Run` functions, rather than spawning subprocesses. This provides significant performance benefits and better integration.

**Execution Flow**:

```go
func (e *CommandExecutor) Execute(commandPath []string, flags map[string]interface{}) (*ExecuteResult, error) {
    // 1. Find the command in the Cobra command tree
    cmd, args, err := e.FindCommand(commandPath)

    // 2. Set up output capture with buffers
    var stdoutBuf, stderrBuf bytes.Buffer

    // 3. Redirect output on root command (since we execute from root)
    originalRootOut := e.rootCmd.OutOrStdout()
    originalRootErr := e.rootCmd.ErrOrStderr()
    e.rootCmd.SetOut(&stdoutBuf)
    e.rootCmd.SetErr(&stderrBuf)

    // 4. Redirect os.Stderr to prevent direct writes from breaking JSON-RPC
    originalOsStderr := os.Stderr
    discardFile, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
    if discardFile != nil {
        os.Stderr = discardFile
        defer func() {
            os.Stderr = originalOsStderr
            discardFile.Close()
        }()
    }

    // 5. Restore original writers after execution
    defer func() {
        e.rootCmd.SetOut(originalRootOut)
        e.rootCmd.SetErr(originalRootErr)
    }()

    // 6. Set flags programmatically on the target command
    for name, value := range flags {
        flag := cmd.Flags().Lookup(name)
        if flag != nil {
            flag.Value.Set(fmt.Sprintf("%v", value))
        }
    }

    // 7. Execute from root command with full path
    ctx := context.Background()
    if cmd != e.rootCmd {
        // Build full command path and execute from root
        fullPath := commandPath
        if len(args) > 0 {
            fullPath = append(fullPath, args...)
        }
        e.rootCmd.SetArgs(fullPath)
        err = e.rootCmd.ExecuteContext(ctx)
    } else {
        cmd.SetArgs(args)
        err = cmd.ExecuteContext(ctx)
    }
    // This calls the command's Run function directly

    // 8. Capture output from buffers
    result := &ExecuteResult{
        Stdout:   stdoutBuf.String(),
        Stderr:   stderrBuf.String(),
        ExitCode: 0,
        Error:    err,
    }

    return result, nil
}
```

**Key Points**:

- **Direct Function Calls**: Commands are executed by calling `rootCmd.ExecuteContext()` with the full command path, which internally calls the command's `Run` function
- **Same Process**: All execution happens in the same process as the MCP server
- **Output Isolation**: Command output is captured via buffers on the root command, preventing interference with the MCP server's stdout/stderr
- **Stderr Protection**: `os.Stderr` is redirected to `/dev/null` to prevent any direct writes from Cobra or other libraries from breaking the JSON-RPC protocol
- **State Management**: Original stdout/stderr are saved and restored to ensure clean state
- **No Subprocess Overhead**: No process spawning, no IPC, no serialization overhead
- **Shared Memory**: Commands share the same memory space, allowing for efficient data access
- **Root Command Execution**: Commands are always executed from the root command with the full path (e.g., `["list", "clusters"]`), not from the leaf command directly

**Benefits**:

- **Performance**: Orders of magnitude faster than subprocess execution
- **Integration**: Commands can access shared state, caches, and connections
- **Debugging**: Easier to debug since everything runs in the same process
- **Resource Efficiency**: No process creation overhead

**Example**:

When `advanced_create` tool is called with `resource="cluster"`:
1. ToolRegistry builds command path: `["create", "cluster"]`
2. CommandExecutor finds the "cluster" command under "create"
3. CommandExecutor redirects `rootCmd.SetOut()` and `rootCmd.SetErr()` to buffers
4. CommandExecutor redirects `os.Stderr` to `/dev/null` to prevent protocol interference
5. CommandExecutor sets args on root command: `rootCmd.SetArgs(["create", "cluster"])`
6. CommandExecutor calls `rootCmd.ExecuteContext(ctx)`
7. Cobra internally calls the command's `Run` function:
   ```go
   Run: func(cmd *cobra.Command, args []string) {
       cmd.Println(`{"id": "cluster-123", "status": "creating"}`)
   }
   ```
8. Output is written to the buffer (via `cmd.SetOut()`)
9. Original stdout/stderr are restored
10. Buffer contents are returned as the result

### 5.5 System Message Generation

```go
func GenerateSystemMessage(config *SystemMessageConfig) string {
    var parts []string

    // Header
    parts = append(parts, fmt.Sprintf("You are a helpful assistant for managing %s resources.", config.CLIName))
    parts = append(parts, "")

    // Tool Usage
    parts = append(parts, "TOOL USAGE:")
    parts = append(parts, fmt.Sprintf("- %s tools use a hierarchical structure...", config.CLIName))
    // ... more instructions

    // Common Patterns
    parts = append(parts, "")
    parts = append(parts, "COMMON PATTERNS:")
    for _, pattern := range config.CommonPatterns {
        parts = append(parts, fmt.Sprintf("- %s", pattern))
    }

    // ... more sections

    return strings.Join(parts, "\n")
}
```

## 6. API Reference

### 6.1 Main Package API

```go
package cobra_mcp

// Server configuration
type ServerConfig struct {
    Name            string
    Version         string
    ToolPrefix      string
    EnableResources bool
    CustomActions   []string  // Custom action names (whitelist mode if set)
    StandaloneCmds  []string  // Commands that don't need resources (whitelist mode if set)
    DangerousCommands []string  // Dangerous commands that require confirmation (format: "action" or "action resource")
}

// Chat configuration
type ChatConfig struct {
    APIKey            string
    APIURL            string
    Model             string
    Debug             bool  // Enable debug output showing tool calls and parameters
    SystemMessage     string  // Optional custom system message (overrides generated)
    SystemMessageFile string  // Optional file path for system message (overrides generated)
    SystemMessageAppend string // Optional content to append to generated system message
}

// System message configuration
type SystemMessageConfig struct {
    CLIName           string
    CLIDescription    string
    ToolPrefix         string
    AvailableActions   []string
    AvailableResources map[string][]string  // action -> resources
    CommonPatterns     []string
    SafetyRequirements []string
    CustomInstructions []string
}

// Create MCP command group (mcp start, mcp stream, mcp tools)
func NewMCPCommand(rootCmd *cobra.Command, config *ServerConfig) *cobra.Command

// Create MCP serve command (deprecated, use NewMCPCommand instead)
func NewMCPServeCommand(rootCmd *cobra.Command, config *ServerConfig) *cobra.Command

// Create chat command
func NewChatCommand(rootCmd *cobra.Command, config *ChatConfig) *cobra.Command

// Generate system message
func GenerateSystemMessage(config *SystemMessageConfig) string

// Low-level API
func NewServer(rootCmd *cobra.Command, config *ServerConfig) *Server
func NewChatClient(server *Server, config *ChatConfig) (*ChatClient, error)
```

### 6.2 Command Line Interface

The library provides MCP commands and a chat command:

**`mcp`** (command group):
- `mcp start` - Start MCP server over stdin/stdout
- `mcp stream` - Start MCP server over HTTP
  - Flags:
    - `--port` (int, default: 8080)
- `mcp tools` - Export available MCP tools as JSON

**`chat`**:
- Flags:
  - `--api-key` (string, or use OPENAI_API_KEY env var)
  - `--api-url` (string, optional custom API URL)
  - `--model` (string, default: gpt-4)
  - `--debug` (bool)
  - `--message` (string, non-interactive mode)
  - `--stdin` (bool, read from stdin)
  - `--system-message-file` (string, custom system message)
- Starts chat client (interactive or non-interactive)

**`chat system-message`**:
- Subcommand to print the system message that would be used for chat
- Useful for debugging and understanding how the AI will be instructed
- Flags:
  - `--system-message-file` (string, custom system message file)

## 7. Example Usage

### 7.1 Basic Integration

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

    // Register your commands
    rootCmd.AddCommand(createCmd)
    rootCmd.AddCommand(listCmd)
    rootCmd.AddCommand(describeCmd)
    rootCmd.AddCommand(deleteCmd)

    // Add MCP commands
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

### 7.2 Custom System Message

**Option 1: Override entire system message**

```go
config := &cobra_mcp.SystemMessageConfig{
    CLIName:        "mycli",
    CLIDescription: "My awesome CLI tool",
    ToolPrefix:     "mycli",
    CommonPatterns: []string{
        "List resources: mycli_list with resource='items'",
        "Create resource: mycli_create with resource='item' and flags={'name': 'my-item'}",
    },
    SafetyRequirements: []string{
        "NEVER execute destructive commands without explicit user confirmation",
        "Always ask for confirmation before deleting resources",
    },
}

systemMessage := cobra_mcp.GenerateSystemMessage(config)
```

**Option 2: Append to generated system message (recommended)**

```go
rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, &cobra_mcp.ChatConfig{
    Model: "gpt-4",
    SystemMessageAppend: `OUTPUT LIMITATION:
When listing resources, always use --parameter size=10 to limit results.
Use --columns to specify only essential columns.`,
}))
```

This approach preserves the auto-generated system message and adds your custom instructions at the end.

### 7.3 Custom Actions

```go
config := &cobra_mcp.ServerConfig{
    Name:          "mycli-mcp-server",
    ToolPrefix:    "mycli",
    CustomActions: []string{"create", "list", "describe", "delete", "apply", "plan"},
    StandaloneCmds: []string{"version", "help", "init"},
}
```

## 8. File Structure

```
cobra-mcp/
├── README.md
├── LICENSE
├── go.mod
├── go.sum
├── cmd/
│   └── example/
│       └── main.go          # Example CLI using the library
├── pkg/
│   ├── executor.go          # CommandExecutor implementation
│   ├── tools.go             # ToolRegistry implementation
│   ├── resources.go         # ResourceRegistry implementation
│   ├── server.go            # MCP Server wrapper
│   ├── chat.go              # ChatClient implementation
│   ├── system_message.go    # System message generation
│   └── config.go            # Configuration types
├── cmd/
│   ├── serve/
│   │   └── cmd.go           # MCP serve command
│   └── chat/
│       └── cmd.go           # Chat command
├── examples/
│   ├── basic/
│   │   └── main.go          # Basic integration example
│   └── advanced/
│       └── main.go          # Advanced customization example
└── tests/
    ├── executor_test.go
    ├── tools_test.go
    ├── server_test.go
    └── chat_test.go
```

## 9. Dependencies

### 9.1 Required Dependencies

- `github.com/spf13/cobra` - Cobra command framework
- `github.com/spf13/pflag` - Flag handling (via Cobra)
- `github.com/modelcontextprotocol/go-sdk/mcp` - MCP protocol implementation
- `github.com/openai/openai-go/v3` - OpenAI API client (for chat)
- `github.com/google/jsonschema-go/jsonschema` - JSON Schema support

### 9.2 Optional Dependencies

- None

## 10. Testing Strategy

### 10.1 Unit Tests

- CommandExecutor: Test command discovery, flag extraction, execution
- ToolRegistry: Test tool generation, hierarchical grouping, tool calling
- ResourceRegistry: Test resource discovery and reading
- SystemMessage: Test message generation with various configs

### 10.2 Integration Tests

- End-to-end MCP server with sample CLI
- Chat client with mock OpenAI API
- Tool execution with real commands

### 10.3 Example CLI

Include a complete example CLI that demonstrates:
- Basic command structure
- MCP server integration
- Chat client integration
- Custom configurations

## 11. Error Handling

### 11.1 Command Execution Errors

- Capture stderr and exit codes
- Return errors in MCP format with `isError: true`
- Provide helpful error messages

### 11.2 Tool Call Errors

- Validate tool names and arguments
- Provide suggestions for hierarchical tools (e.g., "missing resource parameter")
- Return structured error responses

### 11.3 Chat Errors

- Handle API errors gracefully
- Retry logic for transient failures
- Clear error messages for users

## 12. Performance Considerations

### 12.1 Command Discovery

- Cache discovered commands (they don't change at runtime)
- Lazy initialization of tool registry

### 12.2 Tool Execution

- Commands run directly in-process by calling Cobra command Run functions
- Fast execution with no subprocess overhead
- Output is isolated via buffering to avoid interference
- Consider caching for read-only operations (future enhancement)

### 12.3 System Message Generation

- Generate once at startup
- Cache generated messages

## 13. Security Considerations

### 13.1 Command Execution

- Commands run directly in-process by calling Cobra command Run functions
- Output is isolated via buffering (stdout/stderr redirected to buffers)
- No subprocess overhead, but commands share the same process space
- No code injection risks (commands are validated and come from the same application)

### 13.2 API Keys

- Never log API keys
- Support environment variables for API keys
- Validate API keys before use

### 13.3 Destructive Operations

- System messages should warn about destructive operations
- Consider adding confirmation prompts (future enhancement)

## 14. Future Enhancements

1. **Command Caching**: Cache results of read-only commands
2. **Interactive Prompts**: Support interactive command execution
3. **Custom Tool Handlers**: Allow custom tool implementations beyond command execution
4. **Resource Templates**: Template-based resource definitions
5. **Multi-Transport**: Support WebSocket and other transports
6. **Tool Versioning**: Support multiple versions of tools
7. **Metrics**: Add metrics for tool usage
8. **Rate Limiting**: Add rate limiting for tool calls

## 15. Migration Guide

For existing CLIs:

1. Add library dependency
2. Import library
3. Add MCP commands to root command
4. Test with MCP client
5. Customize system message if needed
6. Deploy and test chat functionality

## 16. Conclusion

This design provides a complete, pluggable solution for adding MCP and chat capabilities to any Cobra-based CLI. The hierarchical tool structure reduces complexity while maintaining full functionality. The auto-generated system messages ensure AI assistants understand how to use the tools effectively.

The library is designed to be:
- **Easy to integrate**: Minimal code changes required
- **Flexible**: Highly customizable
- **Robust**: Handles errors gracefully
- **Performant**: Efficient command discovery and execution
- **Secure**: Isolated command execution, safe API key handling
