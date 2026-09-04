// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"bytes"
	"os/exec"
)

func (t *tool) generateCommandLine(args []string) *exec.Cmd {
	var stdout, stderr bytes.Buffer
	cmd := &exec.Cmd{
		Path:   t.exeName,
		Args:   append([]string{t.exeName}, args...),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	return cmd
}
