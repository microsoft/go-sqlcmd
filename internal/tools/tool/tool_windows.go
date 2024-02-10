// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"bytes"
	"fmt"
	"os/exec"
)

func (t *tool) generateCommandLine(args []string) *exec.Cmd {
	var stdout, stderr bytes.Buffer
	fmt.Printf("t.exeName: %v\n", t.exeName)
	fmt.Printf("args: %v\n", args)

	args = append([]string{"foobar"}, args...)

	cmd := &exec.Cmd{
		Path:   t.exeName,
		Args:   args,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	return cmd
}
