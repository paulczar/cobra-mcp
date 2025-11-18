#!/bin/bash
# Minimal test - just check if server responds to initialize

echo "Minimal MCP test - checking if server responds..."
echo "Press Ctrl+C after a few seconds if it hangs..."
echo ""

cd "$(dirname "$0")/.."

# Send initialize and see if we get ANY response
# Use background process to kill after 3 seconds
( sleep 3 && pkill -P $$ 2>/dev/null ) &

echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
go run main.go serve --transport stdio 2>&1 | head -5

kill %1 2>/dev/null || true

echo ""
echo "If you see JSON output above, the server is responding."
echo "If it hangs, the MCP server stdio handling has an issue."
