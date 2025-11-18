package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"

	cobra_mcp "github.com/paulczar/cobra-mcp/pkg"
	"github.com/spf13/cobra"
)

// getTestCLI creates and returns the test CLI root command
func getTestCLI() *cobra.Command {
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
	return rootCmd
}

func TestCmdPrintln(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "cmd-println"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	if !strings.Contains(result.Stdout, "OUTPUT_METHOD:cmd-println") {
		t.Errorf("Expected stdout to contain 'OUTPUT_METHOD:cmd-println', got: %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, `"method": "cmd.Println"`) {
		t.Errorf("Expected stdout to contain cmd.Println JSON, got: %q", result.Stdout)
	}

	if result.Stderr != "" {
		t.Errorf("Expected empty stderr, got: %q", result.Stderr)
	}
}

func TestCmdPrintf(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "cmd-printf"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "OUTPUT_METHOD:cmd-printf") {
		t.Errorf("Expected stdout to contain 'OUTPUT_METHOD:cmd-printf', got: %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, `"value": 42`) {
		t.Errorf("Expected stdout to contain formatted value, got: %q", result.Stdout)
	}
}

func TestFmtPrintln(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "fmt-println"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "OUTPUT_METHOD:fmt-println") {
		t.Errorf("Expected stdout to contain 'OUTPUT_METHOD:fmt-println', got: %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, `"method": "fmt.Println"`) {
		t.Errorf("Expected stdout to contain fmt.Println JSON, got: %q", result.Stdout)
	}
}

func TestFmtPrintf(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "fmt-printf"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "OUTPUT_METHOD:fmt-printf") {
		t.Errorf("Expected stdout to contain 'OUTPUT_METHOD:fmt-printf', got: %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, `"value": 42`) {
		t.Errorf("Expected stdout to contain formatted value, got: %q", result.Stdout)
	}
}

func TestStdoutWrite(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "stdout-write"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "OUTPUT_METHOD:stdout-write") {
		t.Errorf("Expected stdout to contain 'OUTPUT_METHOD:stdout-write', got: %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, `"method": "os.Stdout.Write"`) {
		t.Errorf("Expected stdout to contain os.Stdout.Write JSON, got: %q", result.Stdout)
	}
}

func TestMixedOutput(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "mixed"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	expectedMethods := []string{
		"OUTPUT_METHOD:mixed-1-cmd.Println",
		"OUTPUT_METHOD:mixed-2-fmt.Println",
		"OUTPUT_METHOD:mixed-3-cmd.Printf",
		"OUTPUT_METHOD:mixed-4-fmt.Printf",
		"OUTPUT_METHOD:mixed-5-os.Stdout.Write",
	}

	for _, method := range expectedMethods {
		if !strings.Contains(result.Stdout, method) {
			t.Errorf("Expected stdout to contain %q, got: %q", method, result.Stdout)
		}
	}
}

func TestStderrCapture(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "stderr"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// Check that stdout contains stdout output
	if !strings.Contains(result.Stdout, "STDOUT_METHOD:cmd.Println") {
		t.Errorf("Expected stdout to contain 'STDOUT_METHOD:cmd.Println', got: %q", result.Stdout)
	}

	// Note: stderr is redirected to /dev/null in the executor, so we expect empty stderr
	// The stderr output should not appear in stdout either
	if strings.Contains(result.Stdout, "STDERR_METHOD") {
		t.Errorf("Expected stderr output to be captured separately, but found in stdout: %q", result.Stdout)
	}
}

func TestEmptyOutput(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "empty"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout != "" {
		t.Errorf("Expected empty stdout, got: %q", result.Stdout)
	}

	if result.Stderr != "" {
		t.Errorf("Expected empty stderr, got: %q", result.Stderr)
	}
}

func TestOutputOrdering(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	result, err := executor.Execute([]string{"test", "mixed"}, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// Check that output appears in the correct order
	// Note: Cobra methods (cmd.Println, cmd.Printf) are captured first,
	// then direct writes (fmt.Println, fmt.Printf, os.Stdout.Write) are merged
	stdout := result.Stdout
	idx1 := strings.Index(stdout, "mixed-1") // cmd.Println
	idx2 := strings.Index(stdout, "mixed-2") // fmt.Println
	idx3 := strings.Index(stdout, "mixed-3") // cmd.Printf
	idx4 := strings.Index(stdout, "mixed-4") // fmt.Printf
	idx5 := strings.Index(stdout, "mixed-5") // os.Stdout.Write

	if idx1 == -1 || idx2 == -1 || idx3 == -1 || idx4 == -1 || idx5 == -1 {
		t.Errorf("Not all output methods found in stdout: %q", stdout)
		return
	}

	// Cobra methods come first (1, 3), then direct writes are merged (2, 4, 5)
	// Expected order: cmd.Println (1), cmd.Printf (3), fmt.Println (2), fmt.Printf (4), os.Stdout.Write (5)
	if !(idx1 < idx3 && idx3 < idx2 && idx2 < idx4 && idx4 < idx5) {
		t.Errorf("Output not in expected order. Indices: 1=%d, 2=%d, 3=%d, 4=%d, 5=%d. Output: %q", idx1, idx2, idx3, idx4, idx5, stdout)
	}
}

func TestDangerousCommandExecution(t *testing.T) {
	rootCmd := getTestCLI()
	executor := cobra_mcp.NewCommandExecutor(rootCmd)

	// Test that dangerous commands can still be executed (safety is AI-enforced, not executor-enforced)
	// The command path should match the Cobra command structure: delete -> resource
	result, err := executor.Execute([]string{"delete", "resource"}, map[string]interface{}{
		"name": "test-resource",
	})
	if err != nil {
		// If command not found, check if it's a path issue
		cmd, _, findErr := executor.FindCommand([]string{"delete", "resource"})
		if findErr != nil {
			t.Fatalf("Command not found: %v. FindCommand error: %v", err, findErr)
		}
		if cmd == nil {
			t.Fatalf("Command not found: %v", err)
		}
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %q", result.ExitCode, result.Stderr)
	}

	if !strings.Contains(result.Stdout, `"status": "deleted"`) {
		t.Errorf("Expected stdout to contain deleted status, got: %q", result.Stdout)
	}

	if !strings.Contains(result.Stdout, `"name": "test-resource"`) {
		t.Errorf("Expected stdout to contain resource name, got: %q", result.Stdout)
	}
}

func TestDangerousCommandsInSystemMessage(t *testing.T) {
	rootCmd := getTestCLI()

	// Create server config with dangerous commands
	serverConfig := &cobra_mcp.ServerConfig{
		ToolPrefix:        "testcli",
		DangerousCommands: []string{"delete resource", "delete"},
	}

	server := cobra_mcp.NewServer(rootCmd, serverConfig)

	// Generate system message
	systemMessageConfig := &cobra_mcp.SystemMessageConfig{
		CLIName:           rootCmd.Name(),
		CLIDescription:    rootCmd.Short,
		ToolPrefix:        serverConfig.ToolPrefix,
		DangerousCommands: serverConfig.DangerousCommands,
	}

	systemMessage := cobra_mcp.GenerateSystemMessageFromRegistry(server.ToolRegistry(), rootCmd, systemMessageConfig)

	// Verify dangerous commands are mentioned in system message
	if !strings.Contains(systemMessage, "DANGEROUS COMMANDS") {
		t.Error("Expected system message to contain 'DANGEROUS COMMANDS' section")
	}

	if !strings.Contains(systemMessage, "delete resource") {
		t.Error("Expected system message to contain 'delete resource' in dangerous commands list")
	}

	if !strings.Contains(systemMessage, "delete") {
		t.Error("Expected system message to contain 'delete' in dangerous commands list")
	}

	if !strings.Contains(systemMessage, "CRITICAL SAFETY PROTOCOL") {
		t.Error("Expected system message to contain 'CRITICAL SAFETY PROTOCOL' section")
	}

	if !strings.Contains(systemMessage, "DANGEROUS COMMAND DETECTION") {
		t.Error("Expected system message to contain 'DANGEROUS COMMAND DETECTION' section")
	}
}
