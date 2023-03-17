// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"testing"
)

// TestUninstallWithUserDbPresent installs Mssql
// with a user database, and then uninstalls it using the --force option
func TestUninstallWithUserDbPresent(t *testing.T) {
	const registry = "docker.io"
	const repo = "library/hello-world"

	cmdparser.TestSetup(t)

	cmdparser.TestCmd[*install.Edge](
		fmt.Sprintf(
			`--accept-eula --port 1500 --errorlog-wait-line "Hello from Docker!" --registry %v --repo %v`,
			registry,
			repo))
	cmdparser.TestCmd[*Stop]()
	cmdparser.TestCmd[*Start]()
	cmdparser.TestCmd[*Uninstall]("--yes --force")
}
