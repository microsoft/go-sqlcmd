// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLocalDB(t *testing.T) {
	assert.True(t, isLocalDB(`(localdb)`))
	assert.True(t, isLocalDB(`(LocalDB)\MSSQLLocalDB`))
	assert.True(t, isLocalDB(`(localdb)\.\shared`))
	assert.False(t, isLocalDB(`(localdb)x`))
	assert.False(t, isLocalDB(`localdb`))
	assert.False(t, isLocalDB(`someserver\instance`))
}

func TestSplitServerLocalDB(t *testing.T) {
	tests := []struct {
		input    string
		server   string
		instance string
		protocol string
	}{
		{`(localdb)\MSSQLLocalDB`, `(localdb)`, `MSSQLLocalDB`, ``},
		{`(LocalDb)\.\shared`, `(LocalDb)`, `.\shared`, ``},
		{`np:(localdb)\myinstance`, `(localdb)`, `myinstance`, `np`},
		{`(localdb)`, `(localdb)`, ``, ``},
	}
	for _, test := range tests {
		serverName, instance, port, protocol, err := splitServer(test.input)
		if assert.NoError(t, err, test.input) {
			assert.Equal(t, test.server, serverName, test.input)
			assert.Equal(t, test.instance, instance, test.input)
			assert.Equal(t, uint64(0), port, test.input)
			assert.Equal(t, test.protocol, protocol, test.input)
		}
	}
}

func TestConnectionStringLocalDB(t *testing.T) {
	original := resolveLocalDBPipe
	defer func() { resolveLocalDBPipe = original }()
	var resolved string
	resolveLocalDBPipe = func(instance string) (string, error) {
		resolved = instance
		return `\\.\pipe\LOCALDB#12345678\tsql\query`, nil
	}

	settings := ConnectSettings{ServerName: `(localdb)\MSSQLLocalDB`, Database: "master"}
	connectionString, err := settings.ConnectionString()
	if assert.NoError(t, err) {
		assert.Equal(t, "MSSQLLocalDB", resolved, "instance passed to the LocalDB resolver")
		assert.Equal(t, `sqlserver://.?database=master&pipe=LOCALDB%2312345678%5Ctsql%5Cquery&protocol=np`, connectionString)
	}

	settings = ConnectSettings{ServerName: `tcp:(localdb)\MSSQLLocalDB`}
	_, err = settings.ConnectionString()
	assert.Error(t, err, "LocalDB only supports named pipes")
}
