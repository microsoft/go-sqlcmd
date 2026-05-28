// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import "golang.org/x/sys/windows/registry"

// urlHandlerRegistered reports whether the ssms:// URL handler is registered
// on this machine. SSMS installers (legacy MSI for 18/19/20 and the VS
// Installer for SSMS 21+) register HKEY_CLASSES_ROOT\ssms\shell\open\command,
// so its presence is a reliable install signal regardless of install path.
func (t *SSMS) urlHandlerRegistered() bool {
	k, err := registry.OpenKey(registry.CLASSES_ROOT, `ssms\shell\open\command`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	k.Close()
	return true
}

func (t *SSMS) installText() string {
	return `Install using a package manager:

    winget install Microsoft.SQLServerManagementStudio
    # or
    choco install sql-server-management-studio

Or download the latest version from:

    https://aka.ms/ssmsfullsetup

Note: SSMS is only available on Windows.`
}
