// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

import (
	"testing"
)

func TestCopyToClipboard(t *testing.T) {
	// This test just ensures the function doesn't panic
	// Actual clipboard testing would require platform-specific validation
	err := CopyToClipboard("test password")
	if err != nil {
		// Don't fail on Linux headless environments where clipboard tools may not exist
		t.Logf("CopyToClipboard returned error (may be expected in headless environment): %v", err)
	}
}
