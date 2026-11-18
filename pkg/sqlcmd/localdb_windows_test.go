// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveLocalDBPipe(t *testing.T) {
	if _, err := localDBInstanceAPIPath(); err != nil {
		t.Skip("LocalDB is not installed")
	}
	pipeName, err := resolveLocalDBPipe("")
	if assert.NoError(t, err) {
		assert.True(t, strings.HasPrefix(pipeName, `\\.\pipe\LOCALDB#`), "unexpected pipe name %s", pipeName)
	}
	_, err = resolveLocalDBPipe("no_such_instance_for_sqlcmd_tests")
	assert.Error(t, err, "starting a nonexistent instance should fail")
}

func TestCompareVersions(t *testing.T) {
	assert.Greater(t, compareVersions("16.0", "15.0"), 0)
	assert.Greater(t, compareVersions("110.0", "16.0"), 0)
	assert.Less(t, compareVersions("13.0", "13.1"), 0)
	assert.Equal(t, 0, compareVersions("15.0", "15.0"))
}
