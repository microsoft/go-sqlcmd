// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/open"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// Open defines the `sqlcmd open` sub-commands
type Open struct {
	cmdparser.Cmd
}

func (c *Open) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:         "open",
		Short:       localizer.Sprintf("Open tools (e.g Visual Studio Code, SSMS) for current context"),
		SubCommands: c.SubCommands(),
	}

	c.Cmd.DefineCommand(options)
}

// SubCommands sets up the sub-commands for `sqlcmd open` such as
// `sqlcmd open ads`, `sqlcmd open vscode`, and `sqlcmd open ssms`
func (c *Open) SubCommands() []cmdparser.Command {
	dependencies := c.Dependencies()

	return []cmdparser.Command{
		cmdparser.New[*open.Ads](dependencies),
		cmdparser.New[*open.VSCode](dependencies),
		cmdparser.New[*open.Ssms](dependencies),
	}
}
