// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

func (c *AddUser) encryptPasswordFlag() {
	// Linux OS doesn't have a native DPAPI or Keychain equivalent
}
