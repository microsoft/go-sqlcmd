// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

// CopyToClipboard copies the given text to the system clipboard.
// Returns an error if the clipboard operation fails.
func CopyToClipboard(text string) error {
	return copyToClipboard(text)
}
