// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

// Search in this order
//
//	User Insiders Install
//	System Insiders Install
//	User non-Insiders install
//	System non-Insiders install
func (t *AzureDeveloperCli) searchLocations() []string {

	return []string{"/usr/local/bin/azd"}
}

func (t *AzureDeveloperCli) installText() string {
	return `Install the Azure Developer CLI:

    TODO: Add instructions here

More information can be found here:

    https://learn.microsoft.com/en-us/azure/developer/azure-developer-cli/install-azd`
}
