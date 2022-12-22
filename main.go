// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package main is the entrypoint for the sqlcmd CLI application.
package main

import (
	"os"

	"github.com/microsoft/go-sqlcmd/cmd"
	legacyCmd "github.com/microsoft/go-sqlcmd/cmd/sqlcmd"
)

// main is the entrypoint function for sqlcmd.
//
// TEMPORARY: While we have both the new cobra and old kong CLI
// implementations, main decides which CLI framework to use
func main() {
	cmd.Initialize()

	if isModernCliEnabled() && isFirstArgModernCliSubCommand() {
		cmd.Execute()
	} else {
		legacyCmd.Execute()
	}
}

// isModernCliEnabled is TEMPORARY code, to be removed when we enable
// the new cobra based CLI by default
func isModernCliEnabled() (modernCliEnabled bool) {
	if os.Getenv("SQLCMD_MODERN") != "" {
		modernCliEnabled = true
	}
	return
}

// isFirstArgModernCliSubCommand is TEMPORARY code, to be removed when
// we enable the new cobra based CLI by default
func isFirstArgModernCliSubCommand() (isNewCliCommand bool) {
	if len(os.Args) > 0 {
		if cmd.IsValidSubCommand(os.Args[1]) {
			isNewCliCommand = true
		}
	}
	return
}
