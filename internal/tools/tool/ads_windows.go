// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

// Search in this order
//
//	User Insiders Install
//	System Insiders Install
//	User non-Insiders install
//	System non-Insiders install
func (t *AzureDataStudio) searchLocations() []string {
	userProfile := os.Getenv("USERPROFILE")
	programFiles := os.Getenv("ProgramFiles")

	return []string{
		filepath.Join(userProfile, "AppData\\Local\\Programs\\Azure Data Studio - Insiders\\azuredatastudio-insiders.exe"),
		filepath.Join(programFiles, "Azure Data Studio - Insiders\\azuredatastudio-insiders.exe"),
		filepath.Join(userProfile, "AppData\\Local\\Programs\\Azure Data Studio\\azuredatastudio.exe"),
		filepath.Join(programFiles, "Azure Data Studio\\azuredatastudio.exe"),
	}
}

func (t *AzureDataStudio) installText() string {
	return `Download the latest 'User Installer' .msi from:

    https://go.microsoft.com/fwlink/?linkid=2150927

More information can be found here:

    https://docs.microsoft.com/sql/azure-data-studio/download-azure-data-studio#get-azure-data-studio-for-windows`
}
