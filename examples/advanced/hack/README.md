# MCP Tools Test Script

This directory contains a Python test script to test MCP tools directly without needing ChatGPT.

## Usage

Run the test script:

```bash
python3 test_mcp_tools.py
```

With debug output:

```bash
python3 test_mcp_tools.py --debug
```

## What it tests

The script tests the following MCP tools from the advanced example:

1. **advanced_list** - Lists clusters
2. **advanced_describe** - Describes a cluster
3. **advanced_create** - Creates a cluster
4. **advanced_delete** - Deletes a cluster

## How it works

1. Starts the MCP server as a subprocess (`go run examples/advanced/main.go serve --transport stdio`)
2. Communicates via JSON-RPC over stdio
3. Initializes the MCP connection
4. Lists available tools
5. Calls each tool and verifies the responses

## Requirements

- Python 3.6+
- Go (to build and run the example)
- The advanced example must be buildable

## Example Output

```
Starting MCP server...
Command: go run examples/advanced/main.go serve --transport stdio

Initializing MCP connection...
Initialized: {...}

Sending initialized notification...

Listing available tools...
Found 4 tools:
  - advanced_list: List resources
  - advanced_describe: Describe resources
  - advanced_create: Create resources
  - advanced_delete: Delete resources

Testing advanced_list with resource='clusters'...
Result: {...}
Output: [{"id": "cluster-123", ...}]
✓ Found 2 clusters

...
```

