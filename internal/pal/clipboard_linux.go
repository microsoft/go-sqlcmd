// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"fmt"
	"os/exec"
	"strings"
)

func copyToClipboard(text string) error {
	var attempts []string
	tryCmd := func(name string, args ...string) bool {
		// Use the resolved absolute path so a later PATH change can't redirect
		// the SQL password to a different binary between lookup and exec.
		resolved, err := exec.LookPath(name)
		if err != nil {
			attempts = append(attempts, fmt.Sprintf("%s not found", name))
			return false
		}
		cmd := exec.Command(resolved, args...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			attempts = append(attempts, fmt.Sprintf("%s failed: %v", name, err))
			return false
		}
		return true
	}

	if tryCmd("xclip", "-selection", "clipboard") {
		return nil
	}
	if tryCmd("xsel", "--clipboard", "--input") {
		return nil
	}
	if tryCmd("wl-copy") {
		return nil
	}

	// gotext only scans cmd/modern, cmd/sqlcmd, and pkg/sqlcmd, so this stays plain fmt.Errorf.
	return fmt.Errorf(
		"failed to copy to clipboard; tried xclip, xsel, wl-copy: %s. "+
			"Install one with your distro's package manager (xclip or xsel on X11, wl-clipboard on Wayland) and use your own judgement on which fits your environment",
		strings.Join(attempts, "; "),
	)
}
