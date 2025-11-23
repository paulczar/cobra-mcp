package main

import (
	cobra_mcp "github.com/paulczar/cobra-mcp/pkg"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "example",
		Short: "Example CLI with MCP support",
	}

	// Add some example commands
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources",
	}
	createCmd.AddCommand(&cobra.Command{
		Use:   "cluster",
		Short: "Create a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Creating cluster...")
			return nil
		},
	})
	createCmd.AddCommand(&cobra.Command{
		Use:   "machinepool",
		Short: "Create a machine pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Creating machine pool...")
			return nil
		},
	})

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
	}
	listCmd.AddCommand(&cobra.Command{
		Use:   "clusters",
		Short: "List clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println(`[{"id": "1", "name": "cluster1"}]`)
			return nil
		},
	})

	// Add version command (standalone)
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("example v1.0.0")
			return nil
		},
	}

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)

	// Add MCP commands
	rootCmd.AddCommand(cobra_mcp.NewMCPServeCommand(rootCmd, &cobra_mcp.ServerConfig{
		Name:       "example-mcp-server",
		ToolPrefix: "example",
	}))

	rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, &cobra_mcp.ChatConfig{
		Model: "gpt-5-mini",
	}, nil)) // nil means use default ServerConfig

	rootCmd.Execute()
}
