// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install/edge"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"testing"
)

// TestUninstallWithUserDbPresent(t *testing.T) { installs Mssql (on a specific port to enable parallel testing),
// with a user database, and then uninstalls it using the --force option
func TestUninstallWithUserDbPresent(t *testing.T) {
	const registry = "docker.io"
	const repo = "library/hello-world"

	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*edge.GetTags]()
	cmdparser.TestCmd[*install.Edge](
		fmt.Sprintf(
			`--accept-eula --port-override 1500 --errorlog-wait-line "Hello from Docker!" --registry %v --repo %v`,
			registry,
			repo))
	cmdparser.TestCmd[*Uninstall]("--yes --force")
}
