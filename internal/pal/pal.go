// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package pal provides functions that need to operate differently across different
// operating systems and/or platforms.
package pal

import (
	"os"
	"path/filepath"
	"strings"
)

var lineBreak string

// FilenameInUserHomeDotDirectory returns the full path and filename
// to the filename in the dotDirectory (e.g. .sqlcmd) in the user's home directory
// e.g. c:\users\username
func FilenameInUserHomeDotDirectory(dotDirectory string, filename string) string {
	home, err := os.UserHomeDir()
	checkErr(err)

	return filepath.Join(home, dotDirectory, filename)
}

// UserName returns the name of the currently logged-in user
func UserName() (userName string) {
	return username()
}

// CmdLineWithEnvVars generates a command-line that can be run at the
// operating system command-line (e.g. bash or cmd) which also requires
// one or more environment variables to also be set
func CmdLineWithEnvVars(vars []string, cmd string) string {
	var sb strings.Builder
	for _, v := range vars {
		sb.WriteString(CreateEnvVarKeyword())
		sb.WriteString(cliQuoteIdentifier() + v + cliQuoteIdentifier())
	}
	sb.WriteString(cliCommandSeparator())
	sb.WriteString(cmd)

	return sb.String()
}

func LineBreak() string {
	if lineBreak == "" {
		panic("Initialize has not been called")
	}

	return lineBreak
}
