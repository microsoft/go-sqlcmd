// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build darwin

package sqlcmd

import (
	"os"
	"os/exec"
	"strings"

	"golang.org/x/text/language"
)

// detectUserLocale returns the user's locale from macOS settings
func detectUserLocale() language.Tag {
	// First try environment variables (same as Linux)
	for _, envVar := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if locale := os.Getenv(envVar); locale != "" {
			tag := parseUnixLocale(locale)
			if tag != language.Und {
				return tag
			}
		}
	}

	// Fall back to macOS defaults command
	if locale := getMacOSLocale(); locale != "" {
		if tag, err := language.Parse(locale); err == nil {
			return tag
		}
	}

	return language.English
}

// getMacOSLocale gets the locale from macOS system preferences
func getMacOSLocale() string {
	// Try to get the locale from defaults
	cmd := exec.Command("defaults", "read", "-g", "AppleLocale")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	locale := strings.TrimSpace(string(output))
	// Convert macOS format (en_US) to BCP 47 format (en-US)
	return strings.Replace(locale, "_", "-", -1)
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
	locale = strings.Replace(locale, "_", "-", -1)

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
