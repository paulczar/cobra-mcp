package cobra_mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Server wraps the MCP server implementation
type Server struct {
	rootCmd          *cobra.Command
	toolRegistry     *ToolRegistry
	resourceRegistry *ResourceRegistry
	mcpServer        *mcp.Server
	config           *ServerConfig
}

// NewServer creates a new MCP server
func NewServer(rootCmd *cobra.Command, config *ServerConfig) *Server {
	config = normalizeServerConfig(rootCmd, config)

	// Create registries
	toolRegistry := NewToolRegistry(rootCmd, config)
	var resourceRegistry *ResourceRegistry
	if config.EnableResources {
		executor := NewCommandExecutor(rootCmd)
		resourceRegistry = NewResourceRegistry(executor, config.ToolPrefix)
	}

	// Create MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    config.Name,
		Version: config.Version,
	}, nil)

	server := &Server{
		rootCmd:          rootCmd,
		toolRegistry:     toolRegistry,
		resourceRegistry: resourceRegistry,
		mcpServer:        mcpServer,
		config:           config,
	}

	// Register tools and resources
	server.registerTools()
	if config.EnableResources && resourceRegistry != nil {
		server.registerResources()
	}

	return server
}

// ToolRegistry returns the tool registry
func (s *Server) ToolRegistry() *ToolRegistry {
	return s.toolRegistry
}

// registerTools registers all tools from the ToolRegistry
func (s *Server) registerTools() {
	tools := s.toolRegistry.GetHierarchicalTools()

	for _, toolDef := range tools {
		toolName := toolDef["name"].(string)
		description := toolDef["description"].(string)
		inputSchema := toolDef["inputSchema"].(map[string]interface{})

		// Convert to JSON Schema
		schemaBytes, err := json.Marshal(inputSchema)
		if err != nil {
			continue
		}

		var schema map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &schema); err != nil {
			continue
		}

		// Register tool handler
		s.mcpServer.AddTool(&mcp.Tool{
			Name:        toolName,
			Description: description,
			InputSchema: schema,
		}, func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.handleToolCall(ctx, request)
		})
	}
}

// registerResources registers all resources from the ResourceRegistry
func (s *Server) registerResources() {
	if s.resourceRegistry == nil {
		return
	}

	resources := s.resourceRegistry.GetResources()

	for _, resourceDef := range resources {
		s.mcpServer.AddResource(&mcp.Resource{
			URI:         resourceDef.URI,
			Name:        resourceDef.Name,
			Description: resourceDef.Description,
			MIMEType:    resourceDef.MimeType,
		}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return s.handleReadResource(ctx, request)
		})
	}
}

// handleToolCall handles a tool call request
func (s *Server) handleToolCall(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Convert arguments
	arguments := make(map[string]interface{})
	if len(request.Params.Arguments) > 0 {
		if err := json.Unmarshal(request.Params.Arguments, &arguments); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error parsing arguments: %v", err)},
				},
			}, nil
		}
	}

	// Call tool
	result, err := s.toolRegistry.CallTool(request.Params.Name, arguments)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error calling tool: %v", err)},
			},
		}, nil
	}

	// Convert result to MCP format
	mcpResult := &mcp.CallToolResult{
		IsError: false,
		Content: []mcp.Content{},
	}

	if isError, ok := result["isError"].(bool); ok {
		mcpResult.IsError = isError
	}

	if content, ok := result["content"].([]map[string]interface{}); ok {
		for _, item := range content {
			if text, ok := item["text"].(string); ok {
				mcpResult.Content = append(mcpResult.Content, &mcp.TextContent{Text: text})
			}
		}
	}

	return mcpResult, nil
}

// handleReadResource handles a read resource request
func (s *Server) handleReadResource(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	if s.resourceRegistry == nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     "Resource registry not enabled",
				},
			},
		}, nil
	}

	content, mimeType, err := s.resourceRegistry.ReadResource(request.Params.URI)
	if err != nil {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      request.Params.URI,
					MIMEType: "text/plain",
					Text:     fmt.Sprintf("Error reading resource: %v", err),
				},
			},
		}, nil
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      request.Params.URI,
				MIMEType: mimeType,
				Text:     content,
			},
		},
	}, nil
}

// ServeStdio serves the MCP server over stdio
func ServeStdio(rootCmd *cobra.Command, config *ServerConfig) error {
	server := NewServer(rootCmd, config)

	// Create stdio transport
	transport := &mcp.StdioTransport{}

	// Run server
	return server.mcpServer.Run(context.Background(), transport)
}

// ServeHTTP serves the MCP server over HTTP
func ServeHTTP(rootCmd *cobra.Command, port int, config *ServerConfig) error {
	_ = NewServer(rootCmd, config)

	// Create HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle MCP requests over HTTP
		// This is a simplified implementation
		// In practice, you'd need to handle the MCP protocol over HTTP
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("HTTP transport not fully implemented"))
	})

	// Start HTTP server
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, handler)
}
