// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build darwin

package open

import "os/exec"

// openURI opens a URI via the macOS protocol handler.
func openURI(uri string) error {
	return exec.Command("open", uri).Run()
}
