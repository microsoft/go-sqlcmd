// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"os"
	"os/exec"
	"strings"
)

func copyToClipboard(text string) error {
	cmd := exec.Command(pbcopyPath())
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// pbcopyPath returns the canonical macOS pbcopy path so we never pick up an
// attacker-planted pbcopy from PATH while staging the SQL password on the
// clipboard. Falls back to bare "pbcopy" only if the canonical binary is
// missing (non-standard macOS install).
func pbcopyPath() string {
	const canonical = "/usr/bin/pbcopy"
	if _, err := os.Stat(canonical); err == nil {
		return canonical
	}
	return "pbcopy"
}
