// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

// Search in this order
//
//	User Insiders Install
//	System Insiders Install
//	User non-Insiders install
//	System non-Insiders install
func (t *VisualStudioCode) searchLocations() []string {

	return []string{}
}

func (t *VisualStudioCode) installText() string {
	return `Download the latest installer:

    TODO: Add instructions here

More information can be found here:

    TODO: https://docs.microsoft.com/sql/azure-data-studio/download-azure-data-studio#get-azure-data-studio-for-windows`
}
