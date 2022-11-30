// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import "github.com/spf13/cobra"
import "github.com/microsoft/go-sqlcmd/internal/output"

type AlternativeForFlagInfo struct {
	Flag  string
	Value *string
}

type Cmd struct {
	options Options
	output  output.Output
	command cobra.Command
}

type ExampleInfo struct {
	Description string
	Steps       []string
}
