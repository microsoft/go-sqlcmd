// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import "os"

func envVarCommand() string {
	return "SET"
}

func cliQuoteIdentifier() string {
	return `"`
}

func cliCommandSeparator() string {
	return ` & `
}

func username() string {
	return os.Getenv("USERNAME")
}
