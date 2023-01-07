// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/spf13/cobra"
)

// Command is an interface for defining and running a command which is
// part of a command line program. Command contains methods for setting
// command options, running the command, and checking for errors.
type Command interface {
	// CheckErr checks if the given error is non-nil and, if it is, it prints the error
	// to the output and exits the program with an exit code of 1.
	CheckErr(error)

	// Command returns the underlying cobra.Command object for this command.
	// This is useful for defining subcommands.
	Command() *cobra.Command

	// DefineCommand is used to define a new command and its associated
	// options, flags, and subcommands. It takes in a CommandOptions
	// struct, which allow the caller to specify the command's name, description,
	// usage, and behavior.
	DefineCommand(...CommandOptions)

	// IsSubCommand is TEMPORARY code that will be removed when the
	// old Kong CLI is retired.  It returns true if the command-line
	// provided by the user looks like they want the new cobra CLI, e.g.
	// sqlcmd query, sqlcmd install, sqlcmd --help etc.
	IsSubCommand(command string) bool

	// SetArgsForUnitTesting method allows a caller to set the arguments for the
	// command when running unit tests. This is useful because it allows the caller
	// to simulate different command-line input scenarios in their tests.
	SetArgsForUnitTesting(args []string)

	// SetCrossCuttingConcerns is used to inject cross-cutting concerns (i.e. dependencies)
	// into the Command object (like logging etc.). The dependency.Options allows
	// the Command object to have access to the dependencies it needs, without
	// having to manage them directly.
	SetCrossCuttingConcerns(dependency.Options)
}
