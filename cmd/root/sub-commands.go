// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/root/config"
	"github.com/microsoft/go-sqlcmd/cmd/root/install"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

func SubCommands() []cmdparser.Command {
	return []cmdparser.Command{
		cmdparser.New[*Config](config.SubCommands()...),
		cmdparser.New[*Query](),
		cmdparser.New[*Install](install.SubCommands...),
		cmdparser.New[*Uninstall](),
	}
}
