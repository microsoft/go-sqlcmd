// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package test

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestIsRunningInTestExecutor(t *testing.T) {
	// Test for running in go test on *nix
	os.Args[0] = "main.test"
	assert.True(t, IsRunningInTestExecutor(), "Failed to detect running in go test on *nix")

	// Test for running in go test on windows
	os.Args[0] = "main.test.exe"
	assert.True(t, IsRunningInTestExecutor(), "Failed to detect running in go test on windows")

	// Test for running in goland unittest
	os.Args = []string{"main", "-test.v"}
	assert.True(t, IsRunningInTestExecutor(), "Failed to detect running in goland unittest")

	// Test for not running in test executor
	os.Args = []string{"main"}
	assert.False(t, IsRunningInTestExecutor(), "Incorrectly detected running in test executor")

	// Test for invalid arguments
	os.Args = []string{"main", "-invalid"}
	assert.False(t, IsRunningInTestExecutor(), "Incorrectly detected running in test executor")
}
