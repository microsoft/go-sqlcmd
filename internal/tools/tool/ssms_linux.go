// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

// Search in this order
//
//	User Insiders Install
//	System Insiders Install
//	User non-Insiders install
//	System non-Insiders install
func (t *SqlServerManagementStudio) searchLocations() []string {

	return []string{}
}

func (t *SqlServerManagementStudio) installText() string {
	return `SSMS cannot be installed on this platform.  It is only available on Microsoft Windows`
}
