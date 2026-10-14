// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

package sqlcmd

import "github.com/microsoft/go-sqlcmd/internal/localizer"

func startLocalDBInstance(instance string) (string, error) {
	return "", localizer.Errorf("LocalDB is only supported on Windows")
}
