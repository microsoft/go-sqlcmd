// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import "github.com/microsoft/go-sqlcmd/internal/cmdparser"

func (c *MssqlBase) encryptPasswordFlag(addFlag func(cmdparser.FlagOptions)) {
	addFlag(cmdparser.FlagOptions{
		Bool:  &c.encryptPassword,
		Name:  "encrypt-password",
		Usage: "Encrypt the generated password in the sqlconfig file (using DPAPI)",
	})
}
