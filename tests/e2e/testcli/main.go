package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "testcli",
		Short: "Test CLI for output capture testing",
	}

	// Test cmd.Println
	cmdPrintlnCmd := &cobra.Command{
		Use:   "cmd-println",
		Short: "Test cmd.Println output",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("OUTPUT_METHOD:cmd-println")
			cmd.Println(`{"method": "cmd.Println", "test": true}`)
		},
	}

	// Test cmd.Printf
	cmdPrintfCmd := &cobra.Command{
		Use:   "cmd-printf",
		Short: "Test cmd.Printf output",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("OUTPUT_METHOD:cmd-printf\n")
			cmd.Printf(`{"method": "cmd.Printf", "value": %d}\n`, 42)
		},
	}

	// Test fmt.Println
	fmtPrintlnCmd := &cobra.Command{
		Use:   "fmt-println",
		Short: "Test fmt.Println output",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("OUTPUT_METHOD:fmt-println")
			fmt.Println(`{"method": "fmt.Println", "test": true}`)
		},
	}

	// Test fmt.Printf
	fmtPrintfCmd := &cobra.Command{
		Use:   "fmt-printf",
		Short: "Test fmt.Printf output",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("OUTPUT_METHOD:fmt-printf\n")
			fmt.Printf(`{"method": "fmt.Printf", "value": %d}\n`, 42)
		},
	}

	// Test os.Stdout.Write
	stdoutWriteCmd := &cobra.Command{
		Use:   "stdout-write",
		Short: "Test os.Stdout.Write output",
		Run: func(cmd *cobra.Command, args []string) {
			os.Stdout.Write([]byte("OUTPUT_METHOD:stdout-write\n"))
			os.Stdout.Write([]byte(`{"method": "os.Stdout.Write", "test": true}` + "\n"))
		},
	}

	// Test mixed output methods
	mixedCmd := &cobra.Command{
		Use:   "mixed",
		Short: "Test mixed output methods",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("OUTPUT_METHOD:mixed-1-cmd.Println")
			fmt.Println("OUTPUT_METHOD:mixed-2-fmt.Println")
			cmd.Printf("OUTPUT_METHOD:mixed-3-cmd.Printf\n")
			fmt.Printf("OUTPUT_METHOD:mixed-4-fmt.Printf\n")
			os.Stdout.Write([]byte("OUTPUT_METHOD:mixed-5-os.Stdout.Write\n"))
		},
	}

	// Test stderr capture
	stderrCmd := &cobra.Command{
		Use:   "stderr",
		Short: "Test stderr output",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.PrintErr("STDERR_METHOD:cmd.PrintErr\n")
			fmt.Fprintf(os.Stderr, "STDERR_METHOD:fmt.Fprintf(os.Stderr)\n")
			cmd.Println("STDOUT_METHOD:cmd.Println") // Also output to stdout
		},
	}

	// Test empty output
	emptyCmd := &cobra.Command{
		Use:   "empty",
		Short: "Test empty output",
		Run: func(cmd *cobra.Command, args []string) {
			// Intentionally produce no output
		},
	}

	// Version command (standalone)
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("testcli v1.0.0")
		},
	}

	// Delete command (dangerous command for testing)
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}
	deleteResourceCmd := &cobra.Command{
		Use:   "resource",
		Short: "Delete a resource",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			cmd.Println(fmt.Sprintf(`{"status": "deleted", "name": "%s"}`, name))
		},
	}
	deleteResourceCmd.Flags().String("name", "", "Resource name to delete (required)")
	deleteResourceCmd.MarkFlagRequired("name")
	deleteCmd.AddCommand(deleteResourceCmd)

	// Add all commands
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test commands for output capture",
	}
	testCmd.AddCommand(cmdPrintlnCmd)
	testCmd.AddCommand(cmdPrintfCmd)
	testCmd.AddCommand(fmtPrintlnCmd)
	testCmd.AddCommand(fmtPrintfCmd)
	testCmd.AddCommand(stdoutWriteCmd)
	testCmd.AddCommand(mixedCmd)
	testCmd.AddCommand(stderrCmd)
	testCmd.AddCommand(emptyCmd)

	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
