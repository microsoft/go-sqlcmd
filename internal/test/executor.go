// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package test

import (
	"os"
	"strings"
)

func IsRunningInTestExecutor() bool {
	if strings.HasSuffix(os.Args[0], ".test") || // are we in go test on *nix?
		strings.HasSuffix(os.Args[0], ".test.exe") || // are we in go test on windows?
		(len(os.Args) > 1 && os.Args[1] == "-test.v") { // are we in goland unittest?
		return true
	} else {
		return false
	}
}
