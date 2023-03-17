// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import "os"

func CreateEnvVarKeyword() string {
	return "SET"
}

func cliQuoteIdentifier() string {
	return `"`
}

func cliCommandSeparator() string {
	return ` & `
}

func defaultLineBreak() string {
	return "\n"
}

func username() string {
	return os.Getenv("USERNAME")
}
