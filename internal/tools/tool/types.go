// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

type tool struct {
	installed     *bool
	lookPathError error
	exeName       string
	description   Description
}

type Description struct {
	// Name of the tool, e.g. "Azure Data Studio"
	Name string

	// Purpose describes what this tool does
	Purpose string

	// How to install the tool, e.g. what URL to go to, to download it
	InstallText string
}
