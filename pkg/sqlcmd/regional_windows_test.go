// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build windows

package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLCIDToLanguageTag(t *testing.T) {
	assert.Equal(t, "en-CA", lcidToLanguageTag(0x1009))
	assert.Equal(t, "th-TH", lcidToLanguageTag(0x041e))
	assert.Equal(t, "ar-SA", lcidToLanguageTag(0x0401))
}
