// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package localizer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorfMissingLanguage tests that Errorf returns the correct error message when the language is not found
func TestErrorfMissingLanguage(t *testing.T) {
	os.Setenv("SQLCMD_LANG", "xx-xx")
	err := Errorf("This is a %s error", "test")
	assert.EqualError(t, err, "This is a test error")
}

// TestSprintfPanic tests that Sprintf panics when the language is not found
func TestSprintfPanic(t *testing.T) {
	os.Setenv("SQLCMD_LANG", "")
	assert.Panics(t, func() { Sprintf(nil, "Missing key") })
}
