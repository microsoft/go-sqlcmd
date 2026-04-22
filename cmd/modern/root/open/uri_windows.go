// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"os/exec"
	"syscall"
)

// openURI opens a URI via the Windows shell protocol handler.
func openURI(uri string) error {
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", uri)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
