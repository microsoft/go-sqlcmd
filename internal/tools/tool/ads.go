// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/test"
)

type AzureDataStudio struct {
	tool
}

func (t *AzureDataStudio) Init() {
	t.tool.SetToolDescription(Description{
		Name:        "ads",
		Purpose:     "Azure Data Studio is a database tool for data professionals who use on-premises and cloud data platforms.",
		InstallText: t.installText()})

	for _, location := range t.searchLocations() {
		if file.Exists(location) {
			t.tool.SetExePathAndName(location)
			break
		}
	}
}

func (t *AzureDataStudio) Run(args []string) (int, error) {
	if !test.IsRunningInTestExecutor() {
		return t.tool.Run(args)
	} else {
		return 0, nil
	}
}
