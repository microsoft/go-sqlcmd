// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import "github.com/microsoft/go-sqlcmd/internal/cmdparser"

func (c *MssqlBase) encryptPasswordFlag(addFlag func(options cmdparser.FlagOptions)) {
	// Linux OS doesn't have a native DPAPI or Keychain equivalent
}
