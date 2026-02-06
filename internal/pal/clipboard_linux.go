// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"fmt"
	"os/exec"
	"strings"
)

func copyToClipboard(text string) error {
	// Try xclip first, then xsel, then wl-copy as fallbacks.
	// These are common clipboard utilities on Linux.

	var attempts []string

	// Helper to try a single command and record any errors.
	tryCmd := func(name string, args ...string) bool {
		if _, err := exec.LookPath(name); err != nil {
			attempts = append(attempts, fmt.Sprintf("%s not found", name))
			return false
		}

		cmd := exec.Command(name, args...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			attempts = append(attempts, fmt.Sprintf("%s failed: %v", name, err))
			return false
		}

		return true
	}

	// Try xclip
	if tryCmd("xclip", "-selection", "clipboard") {
		return nil
	}

	// Try xsel as fallback
	if tryCmd("xsel", "--clipboard", "--input") {
		return nil
	}

	// Try wl-copy for Wayland
	if tryCmd("wl-copy") {
		return nil
	}

	// All attempts failed - return combined error message
	return fmt.Errorf("failed to copy to clipboard; tried xclip, xsel, wl-copy: %s", strings.Join(attempts, "; "))
}
