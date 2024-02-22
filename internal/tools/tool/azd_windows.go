// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Search in this order
//
//	User Insiders Install
//	System Insiders Install
//	User non-Insiders install
//	System non-Insiders install
func (t *AzureDeveloperCli) searchLocations() []string {
	userProfile := os.Getenv("USERPROFILE")
	programFiles := os.Getenv("ProgramFiles")

	location, _ := exec.LookPath("azd")

	var locations []string
	if location != "" {
		locations = append(locations, location)
	}

	locations = append(locations, filepath.Join(userProfile, "AppData\\Local\\Programs\\Azure Dev CLI\\azd.exe"))
	locations = append(locations, filepath.Join(programFiles, "Azure Dev CLI\\azd.exe"))

	return locations
}

func (t *AzureDeveloperCli) installText() string {
	return `Install the Azure Developer CLI:

    winget install Microsoft.Azd

More information can be found here:

    https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd`
}
