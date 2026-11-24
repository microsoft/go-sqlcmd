// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	binaryPath string
	buildOnce  sync.Once
	buildErr   error
)

// buildBinary compiles the sqlcmd binary once for all e2e tests.
// The binary is placed in a temporary directory and cleaned up after tests complete.
func buildBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		tmpDir, err := os.MkdirTemp("", "sqlcmd-e2e-*")
		if err != nil {
			buildErr = err
			return
		}
		// Ensure tmpDir is cleaned up if build fails before binaryPath is set
		defer func() {
			if buildErr != nil && binaryPath == "" {
				_ = os.RemoveAll(tmpDir)
			}
		}()

		binaryName := "sqlcmd"
		if runtime.GOOS == "windows" {
			binaryName = "sqlcmd.exe"
		}
		binaryPath = filepath.Join(tmpDir, binaryName)

		cmd := exec.Command("go", "build", "-o", binaryPath, ".")
		// Build from the cmd/modern directory
		wd, err := os.Getwd()
		if err != nil {
			buildErr = err
			return
		}
		cmd.Dir = wd
		output, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = &buildError{err: err, output: string(output)}
			return
		}
	})
	if buildErr != nil {
		t.Fatalf("Failed to build sqlcmd binary: %v", buildErr)
	}
	return binaryPath
}

// hasLiveConnection returns true if SQLCMDSERVER environment variable is set,
// indicating a live SQL Server connection is available for testing.
func hasLiveConnection() bool {
	return os.Getenv("SQLCMDSERVER") != ""
}

// canTestAzureAuth returns true if Azure AD authentication should be used.
// This matches the logic in pkg/sqlcmd/sqlcmd_test.go - use AAD when
// SQLCMDSERVER is an Azure SQL endpoint and SQLCMDUSER is not set.
func canTestAzureAuth() bool {
	server := os.Getenv("SQLCMDSERVER")
	userName := os.Getenv("SQLCMDUSER")
	return strings.Contains(server, ".database.windows.net") && userName == ""
}

// skipIfNoLiveConnection skips the test if no live SQL Server connection is available.
func skipIfNoLiveConnection(t *testing.T) {
	t.Helper()
	if !hasLiveConnection() {
		t.Skip("Skipping: SQLCMDSERVER not set, no live connection available")
	}
}

// getAuthArgs returns the command-line arguments needed for authentication.
// For SQL auth, it returns empty (sqlcmd picks up SQLCMDUSER/SQLCMDPASSWORD from env).
// For Azure AD, it returns the appropriate --authentication-method flag.
func getAuthArgs(t *testing.T) []string {
	t.Helper()
	if canTestAzureAuth() {
		// Check if running in Azure Pipelines with service connection
		if os.Getenv("AZURESUBSCRIPTION_SERVICE_CONNECTION_NAME") != "" {
			t.Log("Using ActiveDirectoryAzurePipelines authentication")
			return []string{"--authentication-method", "ActiveDirectoryAzurePipelines"}
		}
		t.Log("Using ActiveDirectoryDefault authentication")
		return []string{"--authentication-method", "ActiveDirectoryDefault"}
	}
	// SQL auth - credentials come from environment variables
	return []string{}
}

type buildError struct {
	err    error
	output string
}

func (e *buildError) Error() string {
	return e.err.Error() + ": " + e.output
}

// TestE2E_Help verifies that --help flag works and produces expected output.
func TestE2E_Help(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--help")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "sqlcmd --help should not error")
	assert.Contains(t, string(output), "sqlcmd", "help output should mention sqlcmd")
	assert.Contains(t, string(output), "Usage:", "help output should contain Usage section")
}

// TestE2E_Version verifies that --version flag works.
func TestE2E_Version(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--version")
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "sqlcmd --version should not error")
	// Version output should contain version info
	outputStr := string(output)
	assert.True(t, strings.Contains(outputStr, "Version") || strings.Contains(outputStr, "version") || strings.Contains(outputStr, "v"),
		"version output should contain version info: %s", outputStr)
}

// TestE2E_PipedInput_NoPanic verifies that piping input to sqlcmd with -G flag
// does not cause a nil pointer panic. This is a regression test for issue #607.
// The command will fail to connect because it targets a non-existent server, but it should
// NOT panic - that's the key behavior we're testing.
func TestE2E_PipedInput_NoPanic(t *testing.T) {
	binary := buildBinary(t)

	// Create a command that pipes input
	cmd := exec.Command(binary, "-G", "-S", "nonexistent.database.windows.net", "-d", "testdb")
	cmd.Stdin = strings.NewReader("SELECT 1\nGO\n")

	// Run the command - we expect it to fail (can't connect), but NOT panic
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// The command should fail with a connection error, but must not panic
	if err == nil {
		// If it somehow succeeded (unlikely), log the output for debugging
		t.Logf("sqlcmd unexpectedly succeeded: %s", outputStr)
	}

	// Regardless of success or failure, there must be no panic-related output
	assert.NotContains(t, outputStr, "panic:", "sqlcmd should not panic when piping input")
	assert.NotContains(t, outputStr, "nil pointer", "sqlcmd should not have nil pointer error")
	assert.NotContains(t, outputStr, "runtime error", "sqlcmd should not have runtime error")
}

// TestE2E_PipedInput_LiveConnection tests piping input with a real SQL Server connection.
// This test runs when SQLCMDSERVER is set and uses appropriate auth method
// (SQL auth or Azure AD) based on environment configuration.
func TestE2E_PipedInput_LiveConnection(t *testing.T) {
	skipIfNoLiveConnection(t)
	binary := buildBinary(t)

	args := append([]string{"-C"}, getAuthArgs(t)...)
	cmd := exec.Command(binary, args...)
	cmd.Stdin = strings.NewReader("SELECT 1 AS TestValue\nGO\n")
	cmd.Env = os.Environ() // Inherit environment variables

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	require.NoError(t, err, "piped query should succeed with live connection: %s", outputStr)
	assert.Contains(t, outputStr, "TestValue", "output should contain column name")
	assert.Contains(t, outputStr, "1", "output should contain query result")
}

// TestE2E_PipedInput_EmptyInput verifies that piping empty input doesn't panic.
func TestE2E_PipedInput_EmptyInput(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "-S", "nonexistent.server")
	cmd.Stdin = strings.NewReader("")

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Should fail with connection error, but must not panic
	if err != nil {
		t.Logf("Command failed (expected for non-existent server): %v", err)
	}
	assert.NotContains(t, outputStr, "panic:", "sqlcmd should not panic with empty piped input")
	assert.NotContains(t, outputStr, "nil pointer", "sqlcmd should not have nil pointer error")
}

// TestE2E_InvalidFlag verifies that invalid flags produce a helpful error message.
func TestE2E_InvalidFlag(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--this-flag-does-not-exist")
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "invalid flag should cause an error")
	outputStr := string(output)
	// Should have some kind of error message about unknown flag
	assert.True(t, strings.Contains(outputStr, "unknown") || strings.Contains(outputStr, "invalid") || strings.Contains(outputStr, "flag"),
		"error message should indicate unknown/invalid flag: %s", outputStr)
}

// TestE2E_QueryFlag_NoServer verifies -Q flag behavior without a server.
func TestE2E_QueryFlag_NoServer(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "-Q", "SELECT 1")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Should fail because no server is specified, but must not panic
	if err != nil {
		t.Logf("Command failed (expected for no server): %v", err)
	}
	assert.NotContains(t, outputStr, "panic:", "sqlcmd should not panic")
}

// TestE2E_QueryFlag_LiveConnection tests the -Q flag with a real SQL Server connection.
// This test runs when SQLCMDSERVER is set and uses appropriate auth method.
func TestE2E_QueryFlag_LiveConnection(t *testing.T) {
	skipIfNoLiveConnection(t)
	binary := buildBinary(t)

	args := append([]string{"-C", "-Q", "SELECT 42 AS Answer"}, getAuthArgs(t)...)
	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	require.NoError(t, err, "-Q query should succeed: %s", outputStr)
	assert.Contains(t, outputStr, "Answer", "output should contain column name")
	assert.Contains(t, outputStr, "42", "output should contain query result")
}

// TestE2E_InputFile_NotFound verifies proper error when input file doesn't exist.
func TestE2E_InputFile_NotFound(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "-i", "/nonexistent/path/to/file.sql", "-S", "localhost")
	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "non-existent input file should cause an error")
	outputStr := string(output)
	assert.NotContains(t, outputStr, "panic:", "should not panic on missing input file")
}

// TestE2E_InputFile_LiveConnection tests the -i flag with a real SQL Server connection.
// This test runs when SQLCMDSERVER is set and uses appropriate auth method.
func TestE2E_InputFile_LiveConnection(t *testing.T) {
	skipIfNoLiveConnection(t)
	binary := buildBinary(t)

	// Create a temporary SQL file
	tmpFile, err := os.CreateTemp("", "e2e-test-*.sql")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("SELECT 'InputFileTest' AS Source\nGO\n")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	args := append([]string{"-C", "-i", tmpFile.Name()}, getAuthArgs(t)...)
	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	require.NoError(t, err, "-i input file should succeed: %s", outputStr)
	assert.Contains(t, outputStr, "InputFileTest", "output should contain query result from input file")
}

// TestE2E_PipedInput_WithBytesBuffer_NoPanic verifies that piping from bytes.Buffer
// into stdin does not cause a panic, even when the connection fails.
func TestE2E_PipedInput_WithBytesBuffer_NoPanic(t *testing.T) {
	binary := buildBinary(t)

	input := bytes.NewBufferString("SELECT @@VERSION\nGO\n")
	cmd := exec.Command(binary, "-C", "-S", "nonexistent.server")
	cmd.Stdin = input

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command failed (expected for non-existent server): %v", err)
	}
	outputStr := string(output)

	// Should not panic, regardless of whether the connection succeeds or fails
	assert.NotContains(t, outputStr, "panic:", "should not panic when piping SQL with GO")
	assert.NotContains(t, outputStr, "nil pointer", "should not have nil pointer error")
}

func TestE2E_QueryTimeout_NoHang(t *testing.T) {
	skipIfNoLiveConnection(t)
	binary := buildBinary(t)

	args := append([]string{"-C", "-t", "1", "-Q", "WAITFOR DELAY '00:00:10'"}, getAuthArgs(t)...)
	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()

	start := time.Now()
	output, _ := cmd.CombinedOutput()
	elapsed := time.Since(start)

	assert.Contains(t, string(output), "Timeout expired")
	assert.Less(t, elapsed, 30*time.Second, "command hung instead of timing out")
}

// cleanupBinary removes the temporary build directory containing the test binary.
// TestMain calls this to ensure deterministic cleanup instead of relying on
// eventual OS temp directory maintenance.
func cleanupBinary() {
	if binaryPath != "" {
		os.RemoveAll(filepath.Dir(binaryPath))
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	cleanupBinary()
	os.Exit(code)
}
