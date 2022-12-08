// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"github.com/microsoft/go-sqlcmd/cmd/root/install/edge"
	"github.com/microsoft/go-sqlcmd/cmd/root/install/mssql"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

var SubCommands = []cmdparser.Command{
	cmdparser.New[*Mssql](mssql.SubCommands...),
	cmdparser.New[*Edge](edge.SubCommands...),
}
