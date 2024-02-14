// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/test"
	"os"
	"os/exec"
	"runtime"
)

type VisualStudioCode struct {
	tool
}

func (t *VisualStudioCode) Init() {
	t.tool.SetToolDescription(Description{
		Name:        "vscode",
		Purpose:     "Visual Studio Code is a tool for editing files",
		InstallText: t.installText()})

	if runtime.GOOS == "windows" {
		comspec := os.Getenv("COMSPEC")

		t.tool.SetExePathAndName(comspec)
	} else {
		binary, err := exec.LookPath("code")

		if err != nil {
			t.tool.SetExePathAndName(binary)
		}
	}

}

func (t *VisualStudioCode) Run(args []string, options RunOptions) (int, error) {

	if runtime.GOOS == "windows" {
		args = append([]string{"/c", "code"}, args...)
	}

	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args, options)
	} else {
		return 0, nil
	}
}
