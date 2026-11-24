// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows && !linux && !darwin

package sqlcmd

import (
	"os"

	"golang.org/x/text/language"
)

// detectUserLocale returns the user's locale from environment variables.
// This is a fallback implementation for platforms other than Windows, Linux, and Darwin.
// It uses the same environment variable approach as Linux.
func detectUserLocale() language.Tag {
	// Check standard locale environment variables in order of precedence
	for _, envVar := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if locale := os.Getenv(envVar); locale != "" {
			tag := parseUnixLocale(locale)
			if tag != language.Und {
				return tag
			}
		}
	}
	return language.English
}
