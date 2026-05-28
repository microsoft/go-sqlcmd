// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"path/filepath"
)

func (t *SSMS) searchLocations() []string {
	// vswhere is the single source of truth for SSMS 21+ install locations.
	// It finds SSMS wherever it was installed (including non-default drives),
	// which the old hardcoded "Program Files\...18/19/20" list could not.
	// SSMS 20 and earlier (legacy MSI installs) are not detected and are not
	// supported by this command; IsInstalled() will report not-installed and
	// installText() points the user at the latest SSMS.
	root := vswhereFind("Microsoft.VisualStudio.Product.Ssms", t.version)
	if root == "" {
		return nil
	}
	return []string{filepath.Join(root, "Common7", "IDE", "Ssms.exe")}
}

func (t *SSMS) installText() string {
	return `Install the latest version using a package manager:

    winget install Microsoft.SQLServerManagementStudio

Or download the latest version from:

    https://aka.ms/ssmsfullsetup

Note: 'sqlcmd open ssms' supports SSMS 21 and later (discovered via the Visual
Studio Installer). SSMS is only available on Windows.`
}
