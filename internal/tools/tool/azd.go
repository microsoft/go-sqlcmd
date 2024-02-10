// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/test"
)

type AzureDeveloperCli struct {
	tool
}

func (t *AzureDeveloperCli) Init() {
	t.tool.SetToolDescription(Description{
		Name:        "azd",
		Purpose:     "The Azure Developer CLI ( azd ) is a developer-centric command-line interface (CLI) tool for creating Azure applications.",
		InstallText: t.installText()})

	for _, location := range t.searchLocations() {
		if file.Exists(location) {
			t.tool.SetExePathAndName(location)
			break
		}
	}
}

func (t *AzureDeveloperCli) Run(args []string, options RunOptions) (int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args, options)
	} else {
		return 0, nil
	}
}
