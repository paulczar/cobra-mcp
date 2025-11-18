#!/bin/bash
# Simple test to see if MCP server responds to tool calls

echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
go run ../main.go serve --transport stdio 2>&1 | head -5
