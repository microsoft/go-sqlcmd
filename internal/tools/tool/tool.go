// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"fmt"
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
	err := cmd.Run()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return exitCode, err
}

func (t *tool) RunWithOutput(args []string) (string, int, error) {
	if t.installed == nil {
		return "", 1, fmt.Errorf("internal error: Call IsInstalled before RunWithOutput")
	}

	cmd := t.generateCommandLine(args)
	output, err := cmd.Output()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return string(output), exitCode, err
}
