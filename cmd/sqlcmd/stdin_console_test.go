// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"os"
	"testing"

	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/stretchr/testify/assert"
)

func TestIsConsoleInitializationRequiredWithRedirectedStdin(t *testing.T) {
	// Create a temp file to simulate redirected stdin
	tempFile, err := os.CreateTemp("", "stdin-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write some data to it
	_, err = tempFile.WriteString("SELECT 1;\nGO\n")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Remember the original stdin
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Test with a file redirection
	stdinFile, err := os.Open(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer stdinFile.Close()

	// Replace stdin with our redirected file
	os.Stdin = stdinFile

	// Set up a connect settings instance for SQL authentication
	connectConfig := sqlcmd.ConnectSettings{
		UserName: "testuser", // This will trigger SQL authentication, requiring a password
	}

	// Test regular args
	args := &SQLCmdArguments{}

	// Print file stat mode for debugging
	fileStat, _ := os.Stdin.Stat()
	t.Logf("File mode: %v", fileStat.Mode())
	t.Logf("Is character device: %v", (fileStat.Mode()&os.ModeCharDevice) != 0)
	t.Logf("Connection config: %+v", connectConfig)
	t.Logf("RequiresPassword() returns: %v", connectConfig.RequiresPassword())

	// Test with SQL authentication that requires a password
	needsConsole, isInteractive := isConsoleInitializationRequired(&connectConfig, args)
	// Should need console since password is required, but not be interactive
	assert.True(t, needsConsole, "Console should be needed when SQL authentication is used")
	assert.False(t, isInteractive, "Should not be interactive mode with redirected stdin")

	// Now test with no authentication (no password required)
	connectConfig = sqlcmd.ConnectSettings{}
	needsConsole, isInteractive = isConsoleInitializationRequired(&connectConfig, args)
	// Should not need console and not be interactive
	assert.False(t, needsConsole, "Console should not be needed with redirected stdin and no password")
	assert.False(t, isInteractive, "Should not be interactive mode with redirected stdin")

	// Test with direct terminal input (simulated by restoring original stdin)
	os.Stdin = originalStdin
	connectConfig = sqlcmd.ConnectSettings{} // No password needed
	needsConsole, isInteractive = isConsoleInitializationRequired(&connectConfig, args)
	// If no input file or query is specified, it should be interactive mode
	assert.Equal(t, args.InputFile == nil && args.Query == "" && len(args.ChangePasswordAndExit) == 0, needsConsole,
		"Console needs should match interactive mode requirements with terminal stdin")
	assert.Equal(t, args.InputFile == nil && args.Query == "" && len(args.ChangePasswordAndExit) == 0, isInteractive,
		"Interactive mode should be true with terminal stdin and no input files or queries")
}
