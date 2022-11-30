// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/spf13/cobra"
)

type Command interface {
	ArgsForUnitTesting(args []string)
	CheckErr(error)
	Command() *cobra.Command
	DefineCommand(output output.Output, subCommands ...Command)
	Execute()
	Output() output.Output
	SetOptions(Options)
	SetOutput(output.Output)

	// IsSubCommand is TEMPORARY code that will be removed when the
	// new cobra CLI is enabled by default.  It returns true if the command-line
	// provided by the user looks like they want the new cobra CLI, e.g.
	// sqlcmd query, sqlcmd install, sqlcmd --help etc.
	IsSubCommand(command string) bool
}
