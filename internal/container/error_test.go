// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_checkErr(t *testing.T) {
	defer func() { assert.NotNil(t, recover(), "The code did not panic as expected") }()
	checkErr(errors.New("verify error handler"))
}
