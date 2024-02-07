// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/test"
	"os"
)

type VisualStudioCode struct {
	tool
}

func (t *VisualStudioCode) Init() {
	t.tool.SetToolDescription(Description{
		Name:        "vscode",
		Purpose:     "Visual Studio Code is a tool for editing files",
		InstallText: t.installText()})

	// binary, _ := exec.LookPath("code")

	// Get the environment variable COMSPEC
	comspec := os.Getenv("COMSPEC")

	// BUGBUG: This only works on Windows obviously.
	t.tool.SetExePathAndName(comspec)
}

func (t *VisualStudioCode) Run(args []string) (int, error) {
	args = append([]string{"/c", "code"}, args...)

	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args)
	} else {
		return 0, nil
	}
}
