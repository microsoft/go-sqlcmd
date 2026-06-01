// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

// CopyToClipboard copies text to the system clipboard.
func CopyToClipboard(text string) error {
	return copyToClipboard(text)
}
