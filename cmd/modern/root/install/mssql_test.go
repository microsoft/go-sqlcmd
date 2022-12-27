// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/install/mssql"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInstallMssql(t *testing.T) {
	// DEVNOTE: To prevent "import cycle not allowed" golang compile time error (due to
	// cleaning up the Install using root.Uninstall), we don't use root.Uninstall,
	// and use the controller object instead

	cmdparser.TestSetup(t)
	cmdparser.TestCmd[*mssql.GetTags]()
	cmdparser.TestCmd[*Mssql]("--accept-eula --user-database foo")

	controller := container.NewController()
	id := config.ContainerId()
	err := controller.ContainerStop(id)
	assert.Nil(t, err)
	err = controller.ContainerRemove(id)
	assert.Nil(t, err)
}

func TestNegInstallMssql(t *testing.T) {
	assert.Panics(t, func() {

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*Mssql]()
	})
}

func TestNegInstallMssql2(t *testing.T) {
	assert.Panics(t, func() {

		cmdparser.TestSetup(t)
		cmdparser.TestCmd[*Mssql]("--accept-eula --repo does/not/exist")
	})
}
