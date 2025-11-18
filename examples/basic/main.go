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
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Creating cluster...")
		},
	})
	createCmd.AddCommand(&cobra.Command{
		Use:   "machinepool",
		Short: "Create a machine pool",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Creating machine pool...")
		},
	})

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
	}
	listCmd.AddCommand(&cobra.Command{
		Use:   "clusters",
		Short: "List clusters",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(`[{"id": "1", "name": "cluster1"}]`)
		},
	})

	// Add version command (standalone)
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("example v1.0.0")
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
