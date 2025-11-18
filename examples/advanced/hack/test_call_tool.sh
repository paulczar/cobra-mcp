#!/bin/bash
# Test calling an MCP tool

set -e

echo "Testing MCP tool call: advanced_list with resource='clusters'"
echo ""

cd "$(dirname "$0")/.."

# Send requests and capture output
echo "Sending requests..."
{
    # Initialize
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
    sleep 0.3

    # Initialized notification
    echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
    sleep 0.3

    # Call tool
    echo "Calling advanced_list tool..."
    echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"advanced_list","arguments":{"resource":"clusters","flags":{}}}}'

    # Wait longer for tool execution
    sleep 5
} | go run main.go serve --transport stdio 2>&1 | tee /tmp/mcp_tool_call.log

echo ""
echo "--- Checking for response ---"

# Check if we got a response
if grep -q '"id":3' /tmp/mcp_tool_call.log 2>/dev/null; then
    echo "✓ Got tool call response!"
    echo ""
    grep '"id":3' /tmp/mcp_tool_call.log | head -1 | jq '.' 2>/dev/null || grep '"id":3' /tmp/mcp_tool_call.log | head -1
else
    echo "✗ No response for tool call (id=3)"
    echo ""
    echo "Last few lines of output:"
    tail -5 /tmp/mcp_tool_call.log
    echo ""
    echo "This suggests the tool call is hanging."
fi
