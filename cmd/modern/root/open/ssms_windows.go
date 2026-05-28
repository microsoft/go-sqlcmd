// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import "github.com/microsoft/go-sqlcmd/internal/cmdparser"

// Type Ssms is used to implement the "open ssms" which launches SQL Server
// Management Studio and establishes a connection to the SQL Server for the current
// context
type Ssms struct {
	cmdparser.Cmd
}
