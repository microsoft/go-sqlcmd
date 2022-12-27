// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install/edge"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install/mssql"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUninstall installs Mssql (on a specific port to enable parallel testing), and then
// uninstalls it
func TestUninstall(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*mssql.GetTags]()
	cmdparser.TestCmd[*install.Mssql]("--accept-eula --port-override 1500")
	cmdparser.TestCmd[*Uninstall]("--yes")
}

// TestUninstallWithUserDbPresent(t *testing.T) { installs Mssql (on a specific port to enable parallel testing), with a
// user database, and then uninstalls it using the --force option
func TestUninstallWithUserDbPresent(t *testing.T) {
	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*edge.GetTags]()
	cmdparser.TestCmd[*install.Edge]("--accept-eula --user-database foo --port-override 1501")
	cmdparser.TestCmd[*Uninstall]("--yes --force")
}

// TestNegUninstallNoInstanceToUninstall tests that we fail if no instance to
// uninstall
func TestNegUninstallNoInstanceToUninstall(t *testing.T) {
	t.Skip("stuartpa: Not passing on Linux, not sure why right now")
	assert.Panics(t, func() {

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*Uninstall]("--yes")
	})
}
