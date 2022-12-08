// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import "github.com/spf13/cobra"

type Command interface {
	ArgsForUnitTesting(args []string)
	CheckErr(error)
	Command() *cobra.Command
	DefineCommand(subCommands ...Command)
	Execute()

	// IsSubCommand is TEMPORARY code that will be removed when the
	// new cobra CLI is enabled by default.  It returns true if the command-line
	// provided by the user looks like they want the new cobra CLI, e.g.
	// sqlcmd query, sqlcmd install, sqlcmd --help etc.
	IsSubCommand(command string) bool
}
