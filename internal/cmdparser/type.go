// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/spf13/cobra"
)

// Cmd is the main type used for defining and running command line programs.
// It contains fields and methods for defining the command, setting its options,
// and running the command.
type Cmd struct {
	dependencies dependency.Options
	options      CommandOptions
	command      cobra.Command
	unitTesting  bool
}
