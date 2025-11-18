package main

import (
	"encoding/json"

	cobra_mcp "github.com/paulczar/cobra-mcp/pkg"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced example CLI with custom MCP configuration",
		Long:  "This example demonstrates advanced MCP configuration with custom actions and system messages",
	}

	// Add commands with various actions
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources",
	}
	createClusterCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Create a cluster",
		Long:  "Create a new cluster with the specified name, region, and size.",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			region, _ := cmd.Flags().GetString("region")
			size, _ := cmd.Flags().GetString("size")

			// Validate size
			validSizes := map[string]bool{"Small": true, "Medium": true, "Large": true}
			if size != "" && !validSizes[size] {
				cmd.PrintErr("Error: size must be one of: Small, Medium, Large\n")
				return
			}

			// Create response JSON
			response := map[string]interface{}{
				"id":     "cluster-123",
				"name":   name,
				"region": region,
				"size":   size,
				"status": "creating",
			}
			jsonBytes, _ := json.Marshal(response)
			cmd.Println(string(jsonBytes))
		},
	}
	createClusterCmd.Flags().String("name", "", "The unique name of the cluster. The name can be used as the identifier of the cluster. The maximum length is 54 characters. Once set, the cluster name cannot be changed")
	createClusterCmd.Flags().String("region", "", "Use a specific AWS region (such as us-east-1), overriding the AWS_REGION environment variable.")
	createClusterCmd.Flags().String("size", "", "Cluster size: Small, Medium, or Large (required)")
	createClusterCmd.MarkFlagRequired("name")
	createClusterCmd.MarkFlagRequired("region")
	createClusterCmd.MarkFlagRequired("size")
	createCmd.AddCommand(createClusterCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}
	deleteCmd.AddCommand(&cobra.Command{
		Use:   "cluster",
		Short: "Delete a cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(`{"status": "deleted"}`)
		},
	})

	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe resources",
	}
	describeCmd.AddCommand(&cobra.Command{
		Use:   "cluster",
		Short: "Describe a cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(`{"id": "cluster-123", "name": "my-cluster", "status": "ready"}`)
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
			cmd.Println(`[{"id": "cluster-123", "name": "my-cluster", "status": "ready"}, {"id": "cluster-456", "name": "another-cluster", "status": "creating"}]`)
		},
	})

	// Add version command (standalone)
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("advanced v2.0.0")
		},
	}

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)

	// Share the same ServerConfig so dangerous commands work in chat
	serverConfig := &cobra_mcp.ServerConfig{
		Name:       "advanced-mcp-server",
		Version:    "2.0.0",
		ToolPrefix: "advanced",
		// CustomActions:   []string{"create", "delete", "describe", "list"},
		// StandaloneCmds:  []string{"version", "help"},
		EnableResources: true,
		// Dangerous commands that require explicit confirmation
		DangerousCommands: []string{"delete"},
	}

	rootCmd.AddCommand(cobra_mcp.NewMCPServeCommand(rootCmd, serverConfig))
	rootCmd.AddCommand(cobra_mcp.NewChatCommand(rootCmd, &cobra_mcp.ChatConfig{
		Model: "gpt-5-mini",
		Debug: false,
	}, serverConfig))

	rootCmd.Execute()
}
