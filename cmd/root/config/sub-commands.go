// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import "github.com/microsoft/go-sqlcmd/internal/cmdparser"

func SubCommands() []cmdparser.Command {
	return []cmdparser.Command{
		cmdparser.New[*AddContext](),
		cmdparser.New[*AddEndpoint](),
		cmdparser.New[*AddUser](),
		cmdparser.New[*ConnectionStrings](),
		cmdparser.New[*CurrentContext](),
		cmdparser.New[*DeleteContext](),
		cmdparser.New[*DeleteEndpoint](),
		cmdparser.New[*DeleteUser](),
		cmdparser.New[*GetContexts](),
		cmdparser.New[*GetEndpoints](),
		cmdparser.New[*GetUsers](),
		cmdparser.New[*UseContext](),
		cmdparser.New[*View](),
	}
}
