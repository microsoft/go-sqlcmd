// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/io/file"
)

func (t *tool) Init() {
	panic("Do not call directly")
}

func (t *tool) Name() string {
	return t.description.Name
}

func (t *tool) SetExePathAndName(exeName string) {
	t.exeName = exeName
}

// ExePath returns the resolved executable path, or "" if the tool was not
// found. Valid only after Init (and SetBuild, where supported).
func (t *tool) ExePath() string {
	return t.exeName
}

func (t *tool) SetToolDescription(description Description) {
	t.description = description
}

func (t *tool) IsInstalled() bool {
	if t.installed != nil {
		return *t.installed
	}

	t.installed = new(bool)
	// Handle case where tool wasn't found during Init (exeName is empty)
	if t.exeName != "" && file.Exists(t.exeName) {
		*t.installed = true
	} else {
		*t.installed = false
	}

	return *t.installed
}

func (t *tool) HowToInstall() string {
	var sb strings.Builder

	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("%q is not installed on this machine.\n\n", t.description.Name))
	sb.WriteString(fmt.Sprintf("%v\n\n", t.description.Purpose))
	sb.WriteString(fmt.Sprintf("To install %q...\n\n%v\n", t.description.Name, t.description.InstallText))

	return sb.String()
}

func (t *tool) Run(args []string) (int, error) {
	if t.installed == nil {
		return 1, fmt.Errorf("internal error: Call IsInstalled before Run")
	}

	cmd := t.generateCommandLine(args)

	// Redirect stdio to the null device so exec.Cmd does not spawn pipe
	// drainer goroutines. Without this, Start leaves goroutines blocked on
	// the child's stdout/stderr until the GUI tool exits, which keeps
	// sqlcmd's process tree alive even after Process.Release. If opening
	// the null device fails, fall back to inheriting the parent's stdio
	// (also goroutine-free) rather than leaving the bytes.Buffer pipes
	// generateCommandLine attached.
	if devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0); err == nil {
		cmd.Stdin = devNull
		cmd.Stdout = devNull
		cmd.Stderr = devNull
		defer func() { _ = devNull.Close() }()
	} else {
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	// Detach so the launched tool keeps running after sqlcmd exits. GUI tools
	// such as SSMS are the long-running process themselves, so waiting would
	// block until the user closes them.
	return 0, cmd.Process.Release()
}
