// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_checkErr(t *testing.T) {
	assert.Panics(t, func() {
		checkErr(errors.New("verify error handler"))
	})
}
