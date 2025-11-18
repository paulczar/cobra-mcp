#!/usr/bin/env python3
"""
Test script for MCP tools in the advanced example.
Tests calling MCP tools directly without needing ChatGPT.
"""

import json
import subprocess
import sys
import os
import signal
import threading
from typing import Dict, Any, Optional

# Add parent directory to path to find the advanced example
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
EXAMPLE_DIR = os.path.dirname(SCRIPT_DIR)
ROOT_DIR = os.path.dirname(os.path.dirname(os.path.dirname(EXAMPLE_DIR)))


class MCPClient:
    """Simple MCP client that communicates via stdio JSON-RPC."""

    def __init__(self, server_cmd: list, debug: bool = False):
        """Initialize MCP client with server command."""
        self.debug = debug
        self.process = subprocess.Popen(
            server_cmd,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=0
        )
        self.request_id = 1

        if self.debug:
            print(f"[DEBUG] Started process with PID {self.process.pid}")

    def send_request(self, method: str, params: Optional[Dict[str, Any]] = None, timeout: int = 30) -> Dict[str, Any]:
        """Send a JSON-RPC request and return the response."""
        request = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
        }
        if params:
            request["params"] = params

        current_id = self.request_id
        self.request_id += 1

        # Send request
        request_json = json.dumps(request) + "\n"
        if self.debug:
            print(f"[DEBUG] Sending request (id={current_id}): {request_json.strip()}")
        self.process.stdin.write(request_json)
        self.process.stdin.flush()

        # Read response (MCP uses newline-delimited JSON)
        # May need to skip notifications
        import select
        import time

        start_time = time.time()
        while True:
            # Check timeout
            if time.time() - start_time > timeout:
                # Try to read any available output before timing out
                if self.debug:
                    print(f"[DEBUG] Timeout waiting for response. Checking for partial output...")
                    # Try non-blocking read
                    import fcntl
                    try:
                        flags = fcntl.fcntl(self.process.stdout.fileno(), fcntl.F_GETFL)
                        fcntl.fcntl(self.process.stdout.fileno(), fcntl.F_SETFL, flags | os.O_NONBLOCK)
                        partial = self.process.stdout.read()
                        if partial:
                            print(f"[DEBUG] Partial output: {partial}")
                        fcntl.fcntl(self.process.stdout.fileno(), fcntl.F_SETFL, flags)
                    except:
                        pass
                raise Exception(f"Timeout waiting for response to request {current_id} (method: {method})")

            # Check if process died
            if self.process.poll() is not None:
                stderr_output = ""
                try:
                    stderr_output = self.process.stderr.read()
                except:
                    pass
                raise Exception(f"Server process died. Exit code: {self.process.returncode}, stderr: {stderr_output}")

            # Try to read a line (with timeout)
            try:
                # Use select for timeout on read
                import select
                if sys.platform != 'win32':
                    ready, _, _ = select.select([self.process.stdout], [], [], 0.1)
                    if not ready:
                        continue

                response_line = self.process.stdout.readline()
                if not response_line:
                    time.sleep(0.1)
                    continue
            except Exception as e:
                if self.debug:
                    print(f"[DEBUG] Error reading: {e}")
                time.sleep(0.1)
                continue

            if self.debug:
                print(f"[DEBUG] Received line: {response_line.strip()}")

            try:
                response = json.loads(response_line.strip())
            except json.JSONDecodeError as e:
                if self.debug:
                    print(f"[DEBUG] Failed to parse JSON: {e}, line: {response_line}")
                continue

            # Skip notifications (no id field)
            if "id" not in response:
                if self.debug:
                    print(f"[DEBUG] Received notification: {response.get('method', 'unknown')}")
                continue

            # Check if this is the response we're waiting for
            if response.get("id") == current_id:
                if "error" in response:
                    raise Exception(f"Server error: {response['error']}")
                if self.debug:
                    print(f"[DEBUG] Got response: {json.dumps(response, indent=2)}")
                return response.get("result", {})

            # If we got a response with a different ID, log it but keep waiting
            if self.debug:
                print(f"[DEBUG] Received response with different ID: {response.get('id')} (expected {current_id})")

    def send_notification(self, method: str, params: Optional[Dict[str, Any]] = None):
        """Send a JSON-RPC notification (no response expected)."""
        notification = {
            "jsonrpc": "2.0",
            "method": method,
        }
        if params:
            notification["params"] = params

        notification_json = json.dumps(notification) + "\n"
        self.process.stdin.write(notification_json)
        self.process.stdin.flush()

    def initialize(self) -> Dict[str, Any]:
        """Initialize the MCP connection."""
        return self.send_request("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "test-client",
                "version": "1.0.0"
            }
        })

    def list_tools(self) -> Dict[str, Any]:
        """List available tools."""
        return self.send_request("tools/list")

    def call_tool(self, name: str, arguments: Dict[str, Any], timeout: int = 30) -> Dict[str, Any]:
        """Call a tool with the given name and arguments."""
        if self.debug:
            print(f"[DEBUG] Calling tool '{name}' with arguments: {json.dumps(arguments, indent=2)}")
        return self.send_request("tools/call", {
            "name": name,
            "arguments": arguments
        }, timeout=timeout)

    def close(self):
        """Close the connection."""
        if self.process:
            self.process.stdin.close()
            self.process.wait()


def test_list_clusters(client: MCPClient):
    """Test listing clusters."""
    print("Testing advanced_list with resource='clusters'...")
    print("(This may take a moment as the command executes...)")

    try:
        result = client.call_tool("advanced_list", {
            "resource": "clusters",
            "flags": {}
        }, timeout=10)
    except Exception as e:
        print(f"✗ Error calling tool: {e}")
        raise

    print(f"Result: {json.dumps(result, indent=2)}")

    # Verify result structure
    assert "content" in result, "Result should have 'content' field"
    assert not result.get("isError", False), "Should not be an error"

    # Extract text content
    content = result.get("content", [])
    if content and len(content) > 0:
        text_content = content[0].get("text", "")
        print(f"Output: {text_content}")

        # Verify it's valid JSON
        try:
            clusters = json.loads(text_content)
            assert isinstance(clusters, list), "Should return a list of clusters"
            print(f"✓ Found {len(clusters)} clusters")
        except json.JSONDecodeError:
            print("⚠ Output is not valid JSON")
    else:
        print("⚠ No content in response")

    print()


def test_describe_cluster(client: MCPClient):
    """Test describing a cluster."""
    print("Testing advanced_describe with resource='cluster'...")

    result = client.call_tool("advanced_describe", {
        "resource": "cluster",
        "flags": {}
    })

    print(f"Result: {json.dumps(result, indent=2)}")

    assert "content" in result, "Result should have 'content' field"
    assert not result.get("isError", False), "Should not be an error"

    content = result.get("content", [])
    if content and len(content) > 0:
        text_content = content[0].get("text", "")
        print(f"Output: {text_content}")

        try:
            cluster = json.loads(text_content)
            assert isinstance(cluster, dict), "Should return a cluster object"
            print(f"✓ Cluster ID: {cluster.get('id', 'unknown')}")
        except json.JSONDecodeError:
            print("⚠ Output is not valid JSON")

    print()


def test_create_cluster(client: MCPClient):
    """Test creating a cluster."""
    print("Testing advanced_create with resource='cluster'...")

    result = client.call_tool("advanced_create", {
        "resource": "cluster",
        "flags": {}
    })

    print(f"Result: {json.dumps(result, indent=2)}")

    assert "content" in result, "Result should have 'content' field"
    assert not result.get("isError", False), "Should not be an error"

    content = result.get("content", [])
    if content and len(content) > 0:
        text_content = content[0].get("text", "")
        print(f"Output: {text_content}")

        try:
            created = json.loads(text_content)
            assert isinstance(created, dict), "Should return a created object"
            print(f"✓ Created cluster ID: {created.get('id', 'unknown')}")
        except json.JSONDecodeError:
            print("⚠ Output is not valid JSON")

    print()


def test_delete_cluster(client: MCPClient):
    """Test deleting a cluster."""
    print("Testing advanced_delete with resource='cluster'...")

    result = client.call_tool("advanced_delete", {
        "resource": "cluster",
        "flags": {}
    })

    print(f"Result: {json.dumps(result, indent=2)}")

    assert "content" in result, "Result should have 'content' field"
    assert not result.get("isError", False), "Should not be an error"

    content = result.get("content", [])
    if content and len(content) > 0:
        text_content = content[0].get("text", "")
        print(f"Output: {text_content}")

        try:
            deleted = json.loads(text_content)
            print(f"✓ Delete result: {deleted.get('status', 'unknown')}")
        except json.JSONDecodeError:
            print("⚠ Output is not valid JSON")

    print()


def main():
    """Main test function."""
    import argparse

    parser = argparse.ArgumentParser(description="Test MCP tools")
    parser.add_argument("--debug", action="store_true", help="Enable debug output")
    args = parser.parse_args()

    # Build the server command
    # We need to run: go run examples/advanced/main.go serve --transport stdio
    server_cmd = [
        "go", "run",
        os.path.join(EXAMPLE_DIR, "main.go"),
        "serve",
        "--transport", "stdio"
    ]

    print("Starting MCP server...")
    print(f"Command: {' '.join(server_cmd)}")
    print()

    client = None
    try:
        client = MCPClient(server_cmd, debug=args.debug)

        # Initialize connection
        print("Initializing MCP connection...")
        init_result = client.initialize()
        print(f"Initialized: {json.dumps(init_result, indent=2)}")

        # Send initialized notification (required by MCP protocol)
        print("Sending initialized notification...")
        client.send_notification("notifications/initialized")
        print()

        # List available tools
        print("Listing available tools...")
        tools_result = client.list_tools()
        tools = tools_result.get("tools", [])
        print(f"Found {len(tools)} tools:")
        for tool in tools:
            print(f"  - {tool.get('name', 'unknown')}: {tool.get('description', '')}")
        print()

        # Test each tool
        test_list_clusters(client)
        test_describe_cluster(client)
        test_create_cluster(client)
        test_delete_cluster(client)

        print("✓ All tests completed!")

    except Exception as e:
        print(f"✗ Error: {e}", file=sys.stderr)
        if client and client.process:
            stderr = client.process.stderr.read()
            if stderr:
                print(f"Server stderr: {stderr}", file=sys.stderr)
        sys.exit(1)
    finally:
        if client:
            client.close()


if __name__ == "__main__":
    main()
