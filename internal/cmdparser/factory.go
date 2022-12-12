// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/spf13/cobra"
	"os"
)

// Initialize runs the init func() after the command-line provided by the user
// has been parsed.
func Initialize(init func()) {
	cobra.OnInitialize(init)
}

func New[T PtrAsReceiverWrapper[pointerType], pointerType any](dependencies dependency.Options) (command T) {
	if dependencies.Output == nil {
		dependencies.Output = output.New(output.Options{
			OutputType:     "yaml",
			LoggingLevel:   2,
			StandardWriter: os.Stdout,
			ErrorHandler: func(err error) {
				if err != nil {
					panic(err)
				}
			},
			HintHandler: func(hints []string) { fmt.Printf("HINTS: %v\n", hints) }})
	}
	if dependencies.EndOfLine == "" {
		dependencies.EndOfLine = "\n"
	}

	command = new(pointerType)
	command.SetCrossCuttingConcerns(dependencies)
	command.DefineCommand()

	return
}

// PtrAsReceiverWrapper per golang design doc "an unfortunate necessary kludge":
// https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#pointer-method-example
// https://www.reddit.com/r/golang/comments/uqwh5d/generics_new_value_from_pointer_type_with/
type PtrAsReceiverWrapper[T any] interface {
	Command
	*T
}
