// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"bytes"
	"os/exec"
)

func (t *tool) generateCommandLine(args []string) *exec.Cmd {
	// BUGBUG: Move ads specific code to the ads tool
	if tool.Name() == "ads" {
		path, _ := exec.LookPath("open")

		args = append([]string{"--args"}, args...)
		args = append([]string{t.exeName}, args...)
		args = append([]string{"-a"}, args...)
	}

	// BUGBUG: Why is this needed?
	args = append([]string{"."}, args...)

	path := t.exeName
	var stdout, stderr bytes.Buffer
	cmd := &exec.Cmd{
		Path:   path,
		Args:   args,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	return cmd
}
