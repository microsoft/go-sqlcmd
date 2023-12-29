// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/test"
	"os/exec"
)

type VisualStudioCode struct {
	tool
}

func (t *VisualStudioCode) Init() {
	t.tool.SetToolDescription(Description{
		Name:        "vscode",
		Purpose:     "Visual Studio Code is a tool for editing files",
		InstallText: t.installText()})

	binary, _ := exec.LookPath("code")

	fmt.Println("binary: ", binary)
	t.tool.SetExePathAndName(binary)
}

func (t *VisualStudioCode) Run(args []string) (int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args)
	} else {
		return 0, nil
	}
}
