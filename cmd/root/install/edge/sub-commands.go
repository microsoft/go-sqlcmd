// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package edge

import "github.com/microsoft/go-sqlcmd/internal/cmdparser"

var SubCommands = []cmdparser.Command{
	cmdparser.New[*GetTags](),
}
