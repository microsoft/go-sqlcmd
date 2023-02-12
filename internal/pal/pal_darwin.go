// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import "os"

func CreateEnvVarKeyword() string {
	return "export"
}

func cliQuoteIdentifier() string {
	return `'`
}

func cliCommandSeparator() string {
	return `; `
}

func defaultLineBreak() string {
	return "\r\n"
}

func username() string {
	return os.Getenv("USER")
}
