package cobra_mcp

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewMCPServeCommand creates a new MCP serve command
func NewMCPServeCommand(rootCmd *cobra.Command, config *ServerConfig) *cobra.Command {
	var transport string
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long:  "Start the Model Context Protocol server to expose CLI commands as MCP tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			if transport == "stdio" {
				return ServeStdio(rootCmd, config)
			} else if transport == "http" {
				return ServeHTTP(rootCmd, port, config)
			} else {
				return fmt.Errorf("invalid transport: %s (must be 'stdio' or 'http')", transport)
			}
		},
	}

	cmd.Flags().StringVar(&transport, "transport", "stdio", "Transport type: stdio or http")
	cmd.Flags().IntVar(&port, "port", 8080, "Port for HTTP transport (only used when transport=http)")

	return cmd
}

// NewChatCommand creates a new chat command
func NewChatCommand(rootCmd *cobra.Command, config *ChatConfig, serverConfig *ServerConfig) *cobra.Command {
	var apiKey string
	var apiURL string
	var model string
	var debug bool
	var message string
	var stdin bool
	var systemMessageFile string

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an AI chat client with tool calling",
		Long:  "Start an interactive chat client that uses OpenAI API with tool calling to interact with CLI commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use provided server config or create minimal one
			if serverConfig == nil {
				serverConfig = &ServerConfig{
					ToolPrefix: rootCmd.Name(),
				}
			}
			server := NewServer(rootCmd, serverConfig)

			// Create chat config
			chatConfig := config
			if chatConfig == nil {
				chatConfig = &ChatConfig{}
			}

			// Override with flags
			if apiKey != "" {
				chatConfig.APIKey = apiKey
			}
			if apiURL != "" {
				chatConfig.APIURL = apiURL
			}
			if model != "" {
				chatConfig.Model = model
			}
			chatConfig.Debug = debug
			if systemMessageFile != "" {
				chatConfig.SystemMessageFile = systemMessageFile
			}

			// Create chat client
			client, err := NewChatClient(server, chatConfig)
			if err != nil {
				return err
			}

			// Handle different modes
			if message != "" {
				// Single message mode
				return client.ProcessMessage(message)
			} else if stdin {
				// Stdin mode
				return client.ProcessStdin()
			} else {
				// Interactive mode
				return client.RunChatLoop()
			}
		},
	}

	cmd.Flags().StringVar(&apiKey, "api-key", "", "OpenAI API key (or use OPENAI_API_KEY env var)")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Custom API URL (optional)")
	cmd.Flags().StringVar(&model, "model", "gpt-4", "Model to use")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")
	cmd.Flags().StringVar(&message, "message", "", "Single message to process (non-interactive)")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read messages from stdin")
	cmd.Flags().StringVar(&systemMessageFile, "system-message-file", "", "Path to custom system message file")

	// Add system-message subcommand
	cmd.AddCommand(newSystemMessageCommand(rootCmd, config, serverConfig))

	return cmd
}

// newSystemMessageCommand creates a subcommand to print the system message
func newSystemMessageCommand(rootCmd *cobra.Command, config *ChatConfig, serverConfig *ServerConfig) *cobra.Command {
	var systemMessageFile string

	cmd := &cobra.Command{
		Use:   "system-message",
		Short: "Print the system message that would be used for chat",
		Long:  "Print the system message that would be sent to the AI model. This is useful for debugging and understanding how the AI will be instructed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use provided server config or create minimal one
			if serverConfig == nil {
				serverConfig = &ServerConfig{
					ToolPrefix: rootCmd.Name(),
				}
			}
			server := NewServer(rootCmd, serverConfig)

			// Create chat config
			chatConfig := config
			if chatConfig == nil {
				chatConfig = &ChatConfig{}
			}

			// Override with flag
			if systemMessageFile != "" {
				chatConfig.SystemMessageFile = systemMessageFile
			}

			// Get system message (same logic as NewChatClient)
			systemMessage := chatConfig.SystemMessage
			if systemMessage == "" && chatConfig.SystemMessageFile != "" {
				content, err := os.ReadFile(chatConfig.SystemMessageFile)
				if err != nil {
					return fmt.Errorf("error reading system message file: %w", err)
				}
				systemMessage = string(content)
			}

			// Generate system message if not provided
			if systemMessage == "" {
				systemMessageConfig := &SystemMessageConfig{
					CLIName:           rootCmd.Name(),
					CLIDescription:    rootCmd.Short,
					ToolPrefix:        serverConfig.ToolPrefix,
					DangerousCommands: serverConfig.DangerousCommands,
				}
				systemMessage = GenerateSystemMessageFromRegistry(server.toolRegistry, rootCmd, systemMessageConfig)
			}

			// Print the system message
			cmd.Println(systemMessage)
			return nil
		},
	}

	cmd.Flags().StringVar(&systemMessageFile, "system-message-file", "", "Path to custom system message file")

	return cmd
}
