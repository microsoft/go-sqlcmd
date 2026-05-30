// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// copyToClipboard copies text to the Windows clipboard using the built-in clip.exe command.
// This is simpler and safer than using Win32 API calls directly.
func copyToClipboard(text string) error {
	cmd := exec.Command(clipExePath())
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// clipExePath resolves clip.exe under %SystemRoot%\System32 so we never pick
// up an attacker-planted clip.exe from PATH or the working directory while
// copying the SQL password to the clipboard.
func clipExePath() string {
	if root := os.Getenv("SystemRoot"); root != "" {
		return filepath.Join(root, "System32", "clip.exe")
	}
	return "clip.exe"
}
