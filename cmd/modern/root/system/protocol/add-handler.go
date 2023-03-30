// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package protocol

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"runtime"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
)

// AddUser implements the `sqlcmd config add-user` command
type AddHandler struct {
	cmdparser.Cmd
}

func (c *AddHandler) DefineCommand(...cmdparser.CommandOptions) {
	examples := []cmdparser.ExampleOptions{
		{
			Description: "Add a user (using the SQLCMD_PASSWORD environment variable)",
			Steps: []string{
				fmt.Sprintf(`%s SQLCMD_PASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption none",
				fmt.Sprintf(`%s SQLCMD_PASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		},
		{
			Description: "Add a user (using the SQLCMDPASSWORD environment variable)",
			Steps: []string{
				fmt.Sprintf(`%s SQLCMDPASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption none",
				fmt.Sprintf(`%s SQLCMDPASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		},
	}

	if runtime.GOOS == "windows" {
		examples = append(examples, cmdparser.ExampleOptions{
			Description: "Add a user using Windows Data Protection API to encrypt password in sqlconfig",
			Steps: []string{
				fmt.Sprintf(`%s SQLCMD_PASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption dpapi",
				fmt.Sprintf(`%s SQLCMD_PASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		})
	}

	options := cmdparser.CommandOptions{
		Use:      "add-handler",
		Short:    "Add the sqlcmd:// protocol handler",
		Examples: examples,
		Run:      c.run}

	c.Cmd.DefineCommand(options)

}

func (c *AddHandler) run() {
	_ = c.Output()

	/*
		var k registry.Key
		prefix := "SOFTWARE\\Classes\\"
		urlScheme := "sqlcmd"
		basePath := prefix + urlScheme
		permission := uint32(registry.QUERY_VALUE | registry.SET_VALUE)
		baseKey := registry.CURRENT_USER

		programLocation := "\"C:\\Windows\\notepad.exe\""

		// create key
		registry.CreateKey(baseKey, basePath, permission)

		// set description
		k.SetStringValue("", "Notepad app")
		k.SetStringValue("URL Protocol", "")

		// set icon
		registry.CreateKey(registry.CURRENT_USER, "lumiere\\DefaultIcon", registry.ALL_ACCESS)
		k.SetStringValue("", programLocation+",1")

		// create tree
		registry.CreateKey(baseKey, basePath+"\\shell", permission)
		registry.CreateKey(baseKey, basePath+"\\shell\\open", permission)
		registry.CreateKey(baseKey, basePath+"\\shell\\open\\command", permission)

		// set open command
		k.SetStringValue("", programLocation+" \"%1\"")
	*/

}
