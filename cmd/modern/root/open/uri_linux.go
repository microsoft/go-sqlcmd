// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build linux

package open

import "os/exec"

// openURI opens a URI via the Linux protocol handler.
func openURI(uri string) error {
	return exec.Command("xdg-open", uri).Run()
}
