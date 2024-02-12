// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"bytes"
	"fmt"
	"os/exec"
)

func (t *tool) generateCommandLine(args []string) *exec.Cmd {
	//path, _ := exec.LookPath("open")

	// args = append([]string{"--args"}, args...)
	// args = append([]string{t.exeName}, args...)
	// args = append([]string{"-a"}, args...)

	args = append([]string{"foobar"}, args...)

	path := t.exeName
	fmt.Println(path)
	fmt.Println(args)
	var stdout, stderr bytes.Buffer
	cmd := &exec.Cmd{
		Path:   path,
		Args:   args,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	return cmd
}
