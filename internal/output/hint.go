// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

var hintCallback func(hints []string)

func displayHints(hints []string) {
	hintCallback(hints)
}
