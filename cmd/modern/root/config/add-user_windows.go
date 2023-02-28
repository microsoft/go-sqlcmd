// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import "github.com/microsoft/go-sqlcmd/internal/cmdparser"

func (c *AddUser) encryptPasswordFlag() {
	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.encryptPassword,
		Name:  "encrypt-password",
		Usage: "Encrypt the password",
	})
}
