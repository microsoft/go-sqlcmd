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
		// Pin the resolved path so PATH order can't redirect the SQL password between lookup and exec.
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

	return fmt.Errorf(
		"failed to copy to clipboard; tried xclip, xsel, wl-copy: %s\n"+
			"Install one and re-run (use your own judgement on what fits your environment):\n"+
			"  X11:     sudo apt install xclip   (or xsel)\n"+
			"  Wayland: sudo apt install wl-clipboard   (provides wl-copy)\n"+
			"On dnf/pacman-based distros substitute your package manager",
		strings.Join(attempts, "; "),
	)
}
