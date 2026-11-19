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

// clipExePath pins clip.exe to %SystemRoot%\System32 so PATH order can't redirect the SQL password.
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
