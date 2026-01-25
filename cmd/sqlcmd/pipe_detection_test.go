// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"os"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestStdinPipeDetection(t *testing.T) {
	// Get stdin info
	fi, err := os.Stdin.Stat()
	assert.NoError(t, err, "os.Stdin.Stat()")
	
	// On most CI systems, stdin will be a pipe or file (not a terminal)
	// We're testing the logic, not expecting a specific result
	isPipe := false
	if fi != nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		isPipe = true
	}
	
	// Just making sure the detection code doesn't crash
	// The actual value will depend on the environment
	t.Logf("Stdin detected as pipe: %v", isPipe)
}