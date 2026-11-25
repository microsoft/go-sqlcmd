// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"errors"
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

// localDBServerName is the pseudo server name used to address SQL Server Express LocalDB instances
const localDBServerName = "(localdb)"

// defaultLocalDBInstance is the automatic LocalDB instance used when no instance name is given
const defaultLocalDBInstance = "MSSQLLocalDB"

// isLocalDB returns true if serverName is of the form (localdb) or (localdb)\instance
func isLocalDB(serverName string) bool {
	if len(serverName) < len(localDBServerName) || !strings.EqualFold(serverName[:len(localDBServerName)], localDBServerName) {
		return false
	}
	return len(serverName) == len(localDBServerName) || serverName[len(localDBServerName)] == '\\'
}

// resolveLocalDBPipe starts the given LocalDB instance if needed and returns its named pipe
// in the form \\.\pipe\LOCALDB#xxxxxxxx\tsql\query. It is a variable so tests can replace it.
var resolveLocalDBPipe = func(instance string) (string, error) {
	if instance == "" {
		instance = defaultLocalDBInstance
	}
	pipeName, err := startLocalDBInstance(instance)
	if errors.Is(err, errors.ErrUnsupported) {
		return "", localizer.Errorf("LocalDB is only supported on Windows")
	}
	return strings.TrimPrefix(pipeName, "np:"), err
}
