#!/bin/bash
# Simple shell script to test MCP tool calls

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Testing MCP tool: advanced_list with resource='clusters'"
echo ""

# Start the MCP server in the background
echo "Starting MCP server..."
cd "$(dirname "$0")/.."
SERVER_CMD="go run main.go serve --transport stdio"

# Create a named pipe for communication
FIFO=$(mktemp -u)
mkfifo "$FIFO"

# Start server and capture its PID
$SERVER_CMD < "$FIFO" > /tmp/mcp_output.log 2>&1 &
SERVER_PID=$!

# Cleanup function
cleanup() {
    kill $SERVER_PID 2>/dev/null || true
    rm -f "$FIFO"
}
trap cleanup EXIT

# Wait a moment for server to start
sleep 0.5

# Send initialize request
echo "1. Sending initialize request..."
{
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
} > "$FIFO"

# Read response (with timeout)
echo "2. Waiting for initialize response..."
timeout 2 cat < "$FIFO" > /tmp/mcp_response1.log 2>&1 || true
if [ -s /tmp/mcp_response1.log ]; then
    echo -e "${GREEN}✓ Got initialize response:${NC}"
    cat /tmp/mcp_response1.log
else
    echo -e "${RED}✗ No initialize response${NC}"
    cat /tmp/mcp_output.log
    exit 1
fi

# Send initialized notification
echo ""
echo "3. Sending initialized notification..."
{
    echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
} > "$FIFO"

sleep 0.2

# Send tools/list request
echo "4. Listing tools..."
{
    echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
} > "$FIFO"

timeout 2 cat < "$FIFO" > /tmp/mcp_response2.log 2>&1 || true
if [ -s /tmp/mcp_response2.log ]; then
    echo -e "${GREEN}✓ Got tools list:${NC}"
    cat /tmp/mcp_response2.log | head -3
else
    echo -e "${YELLOW}⚠ No tools list response${NC}"
fi

# Send tool call request
echo ""
echo "5. Calling advanced_list tool..."
{
    echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"advanced_list","arguments":{"resource":"clusters","flags":{}}}}'
} > "$FIFO"

echo "   (Waiting up to 5 seconds for response...)"
timeout 5 cat < "$FIFO" > /tmp/mcp_response3.log 2>&1 || true

if [ -s /tmp/mcp_response3.log ]; then
    echo -e "${GREEN}✓ Got tool response:${NC}"
    cat /tmp/mcp_response3.log
    echo ""
    echo -e "${GREEN}✓ Test completed successfully!${NC}"
else
    echo -e "${RED}✗ No tool response (timeout)${NC}"
    echo ""
    echo "Server output:"
    cat /tmp/mcp_output.log
    echo ""
    echo "This suggests the tool call is hanging."
    exit 1
fi
