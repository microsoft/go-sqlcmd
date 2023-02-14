// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

type Description struct {
	Name string

	// Purpose describes what this tool does
	Purpose string

	// Text instructions to install the tool
	InstallText InstallText
}

type InstallText struct {
	Windows string
	Linux   string
	Mac     string
}
