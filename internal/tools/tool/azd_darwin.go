// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

func (t *AzureDeveloperCli) searchLocations() []string {

	return []string{"/usr/local/bin/azd"}
}

func (t *AzureDeveloperCli) installText() string {
	return `Install the Azure Developer CLI:

    brew tap azure/azd && brew install azd

More information can be found here:

    https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd?pivots=os-mac`
}
