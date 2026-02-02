// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package console

import (
	"golang.org/x/term"
	"os"
)

// isStdinRedirected checks if stdin is coming from a pipe or redirection
func isStdinRedirected() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		// If we can't determine, assume it's not redirected
		return false
	}

	// If it's not a character device, it's coming from a pipe or redirection
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return true
	}

	// Double-check using term.IsTerminal
	fd := int(os.Stdin.Fd())
	return !term.IsTerminal(fd)
}
