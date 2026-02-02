// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package console

import (
	"io"
	"os"
	"testing"
)

func TestStdinRedirectionDetection(t *testing.T) {
	// Save original stdin
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	// Create a pipe
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Couldn't create pipe: %v", err)
	}
	defer r.Close()
	defer w.Close()

	// Replace stdin with our pipe
	os.Stdin = r

	// Test if stdin is properly detected as redirected
	if !isStdinRedirected() {
		t.Errorf("Pipe input should be detected as redirected")
	}

	// Write some test input
	go func() {
		_, _ = io.WriteString(w, "test input\n")
		w.Close()
	}()

	// Create console with redirected stdin
	console := NewConsole("")

	// Test readline
	line, err := console.Readline()
	if err != nil {
		t.Fatalf("Failed to read from redirected stdin: %v", err)
	}

	if line != "test input" {
		t.Errorf("Expected 'test input', got '%s'", line)
	}

	// Clean up
	console.Close()
}
