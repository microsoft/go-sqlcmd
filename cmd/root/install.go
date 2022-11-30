// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type Install struct {
	cmdparser.Cmd
}

func (c *Install) DefineCommand(output output.Output, subCommands ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:     "install",
		Short:   "Install/Create #SQLFamily and Tools",
		Aliases: []string{"create"},
	})
	c.Cmd.DefineCommand(
}
