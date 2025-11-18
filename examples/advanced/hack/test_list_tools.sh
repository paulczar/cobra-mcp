#!/bin/bash
# Test script to list available MCP tools

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "Testing MCP tools/list..."
echo ""

cd "$(dirname "$0")/.."

# Create a temporary file for the server output
OUTPUT_FILE=$(mktemp)
trap "rm -f $OUTPUT_FILE" EXIT

# Start server in background
echo "Starting MCP server..."
go run main.go serve --transport stdio > "$OUTPUT_FILE" 2>&1 &
SERVER_PID=$!

# Cleanup function
cleanup() {
    kill $SERVER_PID 2>/dev/null || true
    rm -f "$OUTPUT_FILE"
}
trap cleanup EXIT

# Wait for server to start
sleep 0.5

# Send initialize request
echo "1. Sending initialize request..."
{
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
    sleep 0.2

    # Send initialized notification
    echo "2. Sending initialized notification..."
    echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
    sleep 0.2

    # Send tools/list request
    echo "3. Requesting tools list..."
    echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
    sleep 0.2

    # Give it a moment to respond, then kill
    sleep 1
    kill $SERVER_PID 2>/dev/null || true
} | go run main.go serve --transport stdio 2>&1 | tee /tmp/mcp_tools_output.log

echo ""
echo "--- Server Output ---"
if [ -s /tmp/mcp_tools_output.log ]; then
    echo -e "${GREEN}✓ Got response:${NC}"
    cat /tmp/mcp_tools_output.log | jq '.' 2>/dev/null || cat /tmp/mcp_tools_output.log
    echo ""

    # Try to extract tools
    if command -v jq &> /dev/null; then
        echo -e "${GREEN}Available tools:${NC}"
        cat /tmp/mcp_tools_output.log | jq -r '.result.tools[]?.name // empty' 2>/dev/null | while read tool; do
            if [ -n "$tool" ]; then
                echo "  - $tool"
            fi
        done
    fi
else
    echo -e "${RED}✗ No response received${NC}"
    echo ""
    echo "Server stderr:"
    cat "$OUTPUT_FILE" 2>&1 || true
    exit 1
fi
