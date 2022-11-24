// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

import "github.com/spf13/cobra"

type AlternativeForFlagInfo struct {
	Flag  string
	Value *string
}

type Cmd struct {
	Options Options

	command cobra.Command
}

type ExampleInfo struct {
	Description string
	Steps       []string
}
