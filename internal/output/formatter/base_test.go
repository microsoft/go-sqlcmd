// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBase_CheckErr(t *testing.T) {
	f := &Base{ErrorHandlerCallback: nil}

	assert.Panics(t, func() {
		f.CheckErr(nil)
	})
}
