// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

package open

// run fails immediately on non-Windows platforms. The localized message lives
// in the untagged ssms.go so gotext extracts it regardless of host GOOS.
func (c *SSMS) run() {
	c.Output().Fatal(ssmsUnsupportedPlatformMessage())
}
