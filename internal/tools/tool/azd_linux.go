// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import "os/exec"

func (t *AzureDeveloperCli) searchLocations() []string {
	location, _ := exec.LookPath("azd")

	return []string{location, "/usr/local/bin/azd"}
}

func (t *AzureDeveloperCli) installText() string {
	return `Install the Azure Developer CLI:

    TODO: 

More information can be found here:

    TODO: https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd?pivots=os-mac`
}
