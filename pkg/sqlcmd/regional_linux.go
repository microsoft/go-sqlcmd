// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build linux

package sqlcmd

import (
	"os"
	"strings"

	"golang.org/x/text/language"
)

// detectUserLocale returns the user's locale from environment variables
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

// parseUnixLocale converts a Unix locale string to a language.Tag
// Examples: "en_US.UTF-8", "de_DE", "fr_FR.utf8", "C", "POSIX"
func parseUnixLocale(locale string) language.Tag {
	// Handle special cases
	if locale == "C" || locale == "POSIX" || locale == "" {
		return language.English
	}

	// Remove encoding suffix (e.g., ".UTF-8")
	if idx := strings.Index(locale, "."); idx != -1 {
		locale = locale[:idx]
	}

	// Remove modifier (e.g., "@euro")
	if idx := strings.Index(locale, "@"); idx != -1 {
		locale = locale[:idx]
	}

	// Convert underscore to hyphen for BCP 47 format
	locale = strings.ReplaceAll(locale, "_", "-")

	if tag, err := language.Parse(locale); err == nil {
		return tag
	}

	// Try with just the language part
	if idx := strings.Index(locale, "-"); idx != -1 {
		if tag, err := language.Parse(locale[:idx]); err == nil {
			return tag
		}
	}

	return language.Und
}
