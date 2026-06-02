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

// pbcopyPath pins pbcopy to its canonical path so PATH order can't redirect the SQL password.
func pbcopyPath() string {
	const canonical = "/usr/bin/pbcopy"
	if _, err := os.Stat(canonical); err == nil {
		return canonical
	}
	return "pbcopy"
}
