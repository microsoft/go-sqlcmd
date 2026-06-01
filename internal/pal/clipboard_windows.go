// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func copyToClipboard(text string) error {
	cmd := exec.Command(clipExePath())
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// clipExePath resolves clip.exe under %SystemRoot%\System32 so we never pick
// up an attacker-planted clip.exe from PATH or the working directory while
// copying the SQL password to the clipboard. Falls back through other
// well-known env vars before defaulting to the canonical C:\Windows path so
// we never return a bare "clip.exe" that PATH could resolve.
func clipExePath() string {
	for _, name := range []string{"SystemRoot", "WINDIR"} {
		if v := os.Getenv(name); v != "" {
			return filepath.Join(v, "System32", "clip.exe")
		}
	}
	if drive := os.Getenv("SystemDrive"); drive != "" {
		return filepath.Join(drive+`\`, "Windows", "System32", "clip.exe")
	}
	return `C:\Windows\System32\clip.exe`
}
