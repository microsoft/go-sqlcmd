// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"github.com/spf13/cobra"
)

// Initialize runs the init func() after the command-line provided by the user
// has been parsed.
func Initialize(init func()) {
	cobra.OnInitialize(init)
}

// New creates a cmdparser. After New returns, call Execute() method
// on the top-level Command
//
// Example:
//
//	topLevel : = cmd.New[*MyCommand]()
//	topLevel.Execute()
//
// Example with sub-commands
//
//	topLevel := cmd.New[*MyCommand](MyCommand.subCommands)
func New[T PtrAsReceiverWrapper[CommandPtr], CommandPtr any](subCommands ...Command) (cmd T) {
	cmd = new(CommandPtr)
	cmd.DefineCommand(subCommands...)
	return
}

// PtrAsReceiverWrapper per golang design doc "an unfortunate necessary kludge":
// https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#pointer-method-example
// https://www.reddit.com/r/golang/comments/uqwh5d/generics_new_value_from_pointer_type_with/
type PtrAsReceiverWrapper[T any] interface {
	Command
	*T
}
