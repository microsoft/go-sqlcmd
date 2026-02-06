// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"os/exec"
	"strings"
)

// copyToClipboard copies text to the Windows clipboard using the built-in clip.exe command.
// This is simpler and safer than using Win32 API calls directly.
func copyToClipboard(text string) error {
	cmd := exec.Command("clip")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
