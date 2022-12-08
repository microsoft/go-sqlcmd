// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

type Install struct {
	cmdparser.Cmd
}

func (c *Install) DefineCommand(subCommands ...cmdparser.Command) {
	c.Cmd.Options = cmdparser.Options{
		Use:     "install",
		Short:   "Install/Create #SQLFamily and Tools",
		Aliases: []string{"create"},
	}
	c.Cmd.DefineCommand(subCommands...)
}
