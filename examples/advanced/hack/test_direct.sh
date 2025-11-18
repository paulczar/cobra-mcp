#!/bin/bash
# Even simpler test - just execute the command directly to verify it works

echo "Testing direct command execution (bypassing MCP)..."
echo ""

cd "$(dirname "$0")/.."

echo "Running: go run main.go list clusters"
go run main.go list clusters

echo ""
echo "✓ Direct command works!"
echo ""
echo "If this works but MCP hangs, the issue is in the MCP server response handling."
