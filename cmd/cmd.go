// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"github.com/microsoft/go-sqlcmd/cmd/root"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

func NewRoot() (rootCmd *Root) {
	rootCmd = cmdparser.New[*Root](root.SubCommands()...)
	cmdparser.Initialize(rootCmd.InitializeCallback)
	return
}
