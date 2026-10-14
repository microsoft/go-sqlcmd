// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveLocalDBPipeIsUnsupported(t *testing.T) {
	_, err := resolveLocalDBPipe("MSSQLLocalDB")
	assert.EqualError(t, err, "LocalDB is only supported on Windows")
}
