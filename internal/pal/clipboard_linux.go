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

	if tryCmd("xclip", "-selection", "clipboard") {
		return nil
	}

	if tryCmd("xsel", "--clipboard", "--input") {
		return nil
	}

	if tryCmd("wl-copy") {
		return nil
	}

	return fmt.Errorf("failed to copy to clipboard; tried xclip, xsel, wl-copy: %s", strings.Join(attempts, "; "))
}
