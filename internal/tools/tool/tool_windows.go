// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"bytes"
	"os/exec"
)

func (t *tool) generateCommandLine(args []string) *exec.Cmd {
	var stdout, stderr bytes.Buffer

	// BUGBUG: Why does Cmd ignore the first arg!! (hence I am stuffing it
	// appropriately with 'foobar'
	args = append([]string{"foobar"}, args...)

	cmd := &exec.Cmd{
		Path:   t.exeName,
		Args:   args,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	return cmd
}
