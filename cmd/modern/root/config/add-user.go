// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"os"
	"runtime"

	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/secret"
)

// AddUser implements the `sqlcmd config add-user` command
type AddUser struct {
	cmdparser.Cmd

	name               string
	authType           string
	username           string
	passwordEncryption string
}

func (c *AddUser) DefineCommand(...cmdparser.CommandOptions) {
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
		Use:      "add-user",
		Short:    "Add a user",
		Examples: examples,
		Run:      c.run}

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
		Usage:  "The username (provide password in SQLCMD_PASSWORD or SQLCMDPASSWORD environment variable)",
	})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.passwordEncryption,
		Name:   "password-encryption",
		Usage: fmt.Sprintf("Password encryption method (%s) in sqlconfig file",
			secret.EncryptionMethodsForUsage()),
	})
}

// run a user to the configuration. It sets the user's name and
// authentication type, and, if the authentication type is 'basic', it sets the
// user's username and password (either in plain text or encrypted, depending
// on the --password-encryption flag). If the user's authentication type is not 'basic'
// or 'other', an error is thrown. If the --password-encryption flag is set but the
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

	if c.authType != "basic" && c.passwordEncryption != "" {
		output.FatalWithHints([]string{
			"Remove the --password-encryption flag",
			"Pass in the --auth-type basic"},
			"The --password-encryption flag can only be used when authentication type is 'basic'")
	}

	if c.authType == "basic" && c.passwordEncryption == "" {
		output.FatalWithHints([]string{
			"Add the --password-encryption flag"},
			"The --password-encryption flag must be set when authentication type is 'basic'")
	}

	user := sqlconfig.User{
		Name:               c.name,
		AuthenticationType: c.authType,
	}

	if c.authType == "basic" {
		if os.Getenv("SQLCMD_PASSWORD") == "" && os.Getenv("SQLCMDPASSWORD") == "" {
			output.FatalWithHints([]string{
				"Provide password in the SQLCMD_PASSWORD (or SQLCMDPASSWORD) environment variable"},
				"Authentication Type 'basic' requires a password")
		}

		if c.username == "" {
			output.FatalfWithHintExamples([][]string{
				{"Provide a username with the --username flag",
					"sqlcmd config add-user --username sa"},
			},
				"Username not provided")
		}

		if !secret.IsValidEncryptionMethod(c.passwordEncryption) {
			output.FatalfWithHints([]string{
				fmt.Sprintf("Provide a valid encryption method (%s) with the --password-encryption flag", secret.EncryptionMethodsForUsage())},
				"Encryption method '%v' is not valid", c.passwordEncryption)
		}

		if os.Getenv("SQLCMD_PASSWORD") != "" &&
			os.Getenv("SQLCMDPASSWORD") != "" {
			output.FatalWithHints([]string{
				"Unset one of the environment variables SQLCMD_PASSWORD or SQLCMDPASSWORD"},
				"Both environment variables SQLCMD_PASSWORD and SQLCMDPASSWORD are set. ")
		}

		password := os.Getenv("SQLCMD_PASSWORD")
		if password == "" {
			password = os.Getenv("SQLCMDPASSWORD")
		}
		user.BasicAuth = &sqlconfig.BasicAuthDetails{
			Username:           c.username,
			PasswordEncryption: c.passwordEncryption,
			Password:           secret.Encode(password, c.passwordEncryption),
		}
	}

	uniqueUserName := config.AddUser(user)
	output.Infof("User '%v' added", uniqueUserName)
}
