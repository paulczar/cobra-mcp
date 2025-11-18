#!/bin/bash
# Simpler version - just pipe JSON directly

echo "Testing MCP tools/list (simple version)..."
echo ""

cd "$(dirname "$0")/.."

# Send all requests in sequence
{
    # Initialize
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
    sleep 0.3

    # Initialized notification
    echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
    sleep 0.3

    # List tools
    echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
    sleep 1
} | go run main.go serve --transport stdio 2>&1 | head -20

echo ""
echo "Done. If you see JSON output above with 'tools' array, it worked!"
