// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"os"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/secret"
)

// AddUser implements the `sqlcmd config add-user` command
type AddUser struct {
	cmdparser.Cmd

	name            string
	authType        string
	username        string
	encryptPassword bool
}

func (c *AddUser) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "add-user",
		Short: "Add a user",
		Examples: []cmdparser.ExampleOptions{
			{
				Description: "Add a user",
				Steps: []string{
					fmt.Sprintf(`%s SQLCMD_PASSWORD=<password>`, pal.CreateEnvVarKeyword()),
					"sqlcmd config add-user --name my-user --username user1",
					fmt.Sprintf(`%s SQLCMD_PASSWORD=`, pal.CreateEnvVarKeyword()),
				},
			},
		},
		Run: c.run}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.name,
		Name:          "name",
		DefaultString: "user",
		Usage:         "Display name for the user (this is not the username)",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.authType,
		Name:          "auth-type",
		DefaultString: "basic",
		Usage:         "Authentication type this user will use (basic | other)",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.username,
		Name:   "username",
		Usage:  "The username (provide password in SQLCMD_PASSWORD environment variable)",
	})

	c.encryptPasswordFlag()
}

// run a user to the configuration. It sets the user's name and
// authentication type, and, if the authentication type is 'basic', it sets the
// user's username and password (either in plain text or encrypted, depending
// on the --encrypt-password flag). If the user's authentication type is not 'basic'
// or 'other', an error is thrown. If the --encrypt-password flag is set but the
// authentication type is not 'basic', an error is thrown. If the authentication
// type is 'basic' but the username or password is not provided, an error is thrown.
// If the username is provided but the password is not, an error is thrown.
func (c *AddUser) run() {
	output := c.Output()

	if c.authType != "basic" &&
		c.authType != "other" {
		output.FatalfWithHints([]string{"Authentication type must be 'basic' or 'other'"},
			"Authentication type '' is not valid %v'", c.authType)
	}

	if c.authType != "basic" && c.encryptPassword {
		output.FatalWithHints([]string{
			"Remove the --encrypt-password flag",
			"Pass in the --auth-type basic"},
			"The --encrypt-password flag can only be used when authentication type is 'basic'")
	}

	user := sqlconfig.User{
		Name:               c.name,
		AuthenticationType: c.authType,
	}

	if c.authType == "basic" {
		if os.Getenv("SQLCMD_PASSWORD") == "" {
			output.FatalWithHints([]string{
				"Provide password in the SQLCMD_PASSWORD environment variable"},
				"Authentication Type 'basic' requires a password")
		}

		if c.username == "" {
			output.FatalfWithHintExamples([][]string{
				{"Provide a username with the --username flag",
					"sqlcmd config add-user --username stuartpa"},
			},
				"Username not provider")
		}

		user.BasicAuth = &sqlconfig.BasicAuthDetails{
			Username:          c.username,
			PasswordEncrypted: c.encryptPassword,
			Password:          secret.Encode(os.Getenv("SQLCMD_PASSWORD"), c.encryptPassword),
		}
	}

	config.AddUser(user)
	output.Infof("User '%v' added", user.Name)
}
