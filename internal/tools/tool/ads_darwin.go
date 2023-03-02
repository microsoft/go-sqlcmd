// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
)

func (t *AzureDataStudio) searchLocations() []string {
	userProfile := os.Getenv("HOME")

	return []string{
		filepath.Join("/", "Applications", "Azure Data Studio - Insiders.app"),
		filepath.Join(userProfile, "Downloads", "Azure Data Studio - Insiders.app"),
		filepath.Join("/", "Applications", "Azure Data Studio.app"),
		filepath.Join(userProfile, "Downloads", "Azure Data Studio.app"),
	}
}

func (t *AzureDataStudio) installText() string {
	return `Download the latest .zip from:

    https://go.microsoft.com/fwlink/?linkid=2151311

More information can be found here:

    https://docs.microsoft.com/sql/azure-data-studio/download-azure-data-studio?#get-azure-data-studio-for-macos`
}
