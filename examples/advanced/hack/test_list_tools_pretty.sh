#!/bin/bash
# Pretty version that formats the tools list nicely

echo "Testing MCP tools/list..."
echo ""

cd "$(dirname "$0")/.."

# Send requests and capture output
OUTPUT=$({
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
    sleep 0.3
    echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
    sleep 0.3
    echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
    sleep 1
} | go run main.go serve --transport stdio 2>&1)

# Extract tools list response
TOOLS_RESPONSE=$(echo "$OUTPUT" | grep '"method":"tools/list"' -A 100 | grep '"id":2' | head -1)

if [ -z "$TOOLS_RESPONSE" ]; then
    # Try to find the tools/list response by ID
    TOOLS_RESPONSE=$(echo "$OUTPUT" | grep '"id":2' | head -1)
fi

if [ -n "$TOOLS_RESPONSE" ]; then
    echo "✓ Successfully retrieved tools list!"
    echo ""

    if command -v jq &> /dev/null; then
        echo "Available tools:"
        echo "$TOOLS_RESPONSE" | jq -r '.result.tools[]? | "  - \(.name): \(.description)"' 2>/dev/null || {
            echo "$TOOLS_RESPONSE" | jq '.result.tools[]?.name' 2>/dev/null | sed 's/"//g' | sed 's/^/  - /'
        }
    else
        echo "Tools found (install 'jq' for better formatting):"
        echo "$TOOLS_RESPONSE" | grep -o '"name":"[^"]*"' | sed 's/"name":"/  - /' | sed 's/"$//'
    fi
else
    echo "✗ Could not find tools list in response"
    echo ""
    echo "Full output:"
    echo "$OUTPUT"
fi
