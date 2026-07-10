// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/test"
)

type VSCode struct {
	tool

	// build pins which VS Code build to discover and launch: "stable",
	// "insiders", or "" for the default (stable first, then insiders).
	build string
}

func (t *VSCode) Init() {
	t.SetToolDescription(Description{
		Name:        "vscode",
		Purpose:     "Visual Studio Code is a code editor with support for database management through the MSSQL extension.",
		InstallText: t.installText()})

	t.resolveExePath()
}

// SetBuild pins the VS Code build to discover and re-resolves the exe path.
// Call after NewTool and before IsInstalled.
func (t *VSCode) SetBuild(build string) {
	t.build = strings.ToLower(build)
	t.resolveExePath()
}

func (t *VSCode) resolveExePath() {
	t.exeName = ""
	t.installed = nil
	for _, location := range t.searchLocations() {
		if file.Exists(location) {
			t.SetExePathAndName(location)
			break
		}
	}
}

// buildsToSearch returns the build identifiers to probe, in priority order.
// An unset build defaults to stable first, then insiders.
func (t *VSCode) buildsToSearch() []string {
	switch t.build {
	case "stable":
		return []string{"stable"}
	case "insiders":
		return []string{"insiders"}
	default:
		return []string{"stable", "insiders"}
	}
}

func (t *VSCode) Run(args []string) (int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.launch(args)
	}
	return 0, nil
}
