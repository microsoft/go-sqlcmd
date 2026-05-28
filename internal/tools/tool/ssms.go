// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

type SSMS struct {
	tool
}

func (t *SSMS) Init() {
	t.SetToolDescription(Description{
		Name:        "ssms",
		Purpose:     "SQL Server Management Studio (SSMS) is an integrated environment for managing SQL Server infrastructure.",
		InstallText: t.installText()})
}

// IsInstalled reports whether SSMS is installed by checking for the ssms://
// URL handler registration. Launch is performed via that URL handler in
// cmd/modern/root/open, so Run is not implemented for SSMS.
func (t *SSMS) IsInstalled() bool {
	return t.urlHandlerRegistered()
}
