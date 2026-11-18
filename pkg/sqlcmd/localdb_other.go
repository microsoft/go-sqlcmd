// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

package sqlcmd

import "errors"

func startLocalDBInstance(instance string) (string, error) {
	return "", errors.ErrUnsupported
}
