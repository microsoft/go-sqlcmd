// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/telemetry"

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
			Description: localizer.Sprintf("Add a user (using the SQLCMD_PASSWORD environment variable)"),
			Steps: []string{
				fmt.Sprintf(`%s SQLCMD_PASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption none",
				fmt.Sprintf(`%s SQLCMD_PASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		},
		{
			Description: localizer.Sprintf("Add a user (using the SQLCMDPASSWORD environment variable)"),
			Steps: []string{
				fmt.Sprintf(`%s SQLCMDPASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption none",
				fmt.Sprintf(`%s SQLCMDPASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		},
	}

	if runtime.GOOS == "windows" {
		examples = append(examples, cmdparser.ExampleOptions{
			Description: localizer.Sprintf("Add a user using Windows Data Protection API to encrypt password in sqlconfig"),
			Steps: []string{
				fmt.Sprintf(`%s SQLCMD_PASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption dpapi",
				fmt.Sprintf(`%s SQLCMD_PASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		})
	}

	options := cmdparser.CommandOptions{
		Use:      "add-user",
		Short:    localizer.Sprintf("Add a user"),
		Examples: examples,
		Run:      c.run}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.name,
		Name:          "name",
		DefaultString: "user",
		Usage:         localizer.Sprintf("Display name for the user (this is not the username)"),
	})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.authType,
		Name:          "auth-type",
		DefaultString: "basic",
		Usage:         localizer.Sprintf("Authentication type this user will use (basic | other)"),
	})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.username,
		Name:   "username",
		Usage:  localizer.Sprintf("The username (provide password in %s or %s environment variable)", localizer.PasswordEnvVar, localizer.PasswordEnvVar2),
	})

	c.AddFlag(cmdparser.FlagOptions{
		String: &c.passwordEncryption,
		Name:   "password-encryption",
		Usage:  localizer.Sprintf("Password encryption method (%s) in sqlconfig file", secret.EncryptionMethodsForUsage()),
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
		output.FatalWithHints([]string{localizer.Sprintf("Authentication type must be '%s' or '%s'", localizer.ModernAuthTypeBasic, localizer.ModernAuthTypeOther)},
			localizer.Sprintf("Authentication type '' is not valid %v'", c.authType))
	}

	if c.authType != "basic" && c.passwordEncryption != "" {
		output.FatalWithHints([]string{
			localizer.Sprintf("Remove the %s flag", localizer.PasswordEncryptFlag),
			localizer.Sprintf("Pass in the %s %s", localizer.AuthTypeFlag, localizer.ModernAuthTypeBasic)},
			localizer.Sprintf("The %s flag can only be used when authentication type is '%s'", localizer.PasswordEncryptFlag, localizer.ModernAuthTypeBasic))
	}

	if c.authType == "basic" && c.passwordEncryption == "" {
		output.FatalWithHints([]string{
			localizer.Sprintf("Add the %s flag", localizer.PasswordEncryptFlag)},
			localizer.Sprintf("The %s flag must be set when authentication type is '%s'", localizer.PasswordEncryptFlag, localizer.ModernAuthTypeBasic))
	}

	user := sqlconfig.User{
		Name:               c.name,
		AuthenticationType: c.authType,
	}

	if c.authType == "basic" {
		if os.Getenv("SQLCMD_PASSWORD") == "" && os.Getenv("SQLCMDPASSWORD") == "" {
			output.FatalWithHints([]string{
				localizer.Sprintf("Provide password in the %s (or %s) environment variable", localizer.PasswordEnvVar, localizer.PasswordEnvVar2)},
				localizer.Sprintf("Authentication Type '%s' requires a password", localizer.ModernAuthTypeBasic))
		}

		if c.username == "" {
			output.FatalWithHintExamples([][]string{
				{localizer.Sprintf("Provide a username with the %s flag"),
					"sqlcmd config add-user --username sa"},
			},
				localizer.Sprintf("Username not provided"))
		}

		if !secret.IsValidEncryptionMethod(c.passwordEncryption) {
			output.FatalWithHints([]string{
				localizer.Sprintf("Provide a valid encryption method (%s) with the %s flag", secret.EncryptionMethodsForUsage(), localizer.PasswordEncryptFlag)},
				localizer.Sprintf("Encryption method '%v' is not valid", c.passwordEncryption))
		}

		if os.Getenv("SQLCMD_PASSWORD") != "" &&
			os.Getenv("SQLCMDPASSWORD") != "" {
			output.FatalWithHints([]string{
				localizer.Sprintf("Unset one of the environment variables %s or %s", localizer.PasswordEnvVar, localizer.PasswordEnvVar2)},
				localizer.Sprintf("Both environment variables %s and %s are set. ", localizer.PasswordEnvVar, localizer.PasswordEnvVar2))
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
	output.Info(localizer.Sprintf("User '%v' added", uniqueUserName))
}

func (c *AddUser) LogTelemtry() {
	eventName := "config-add-user"
	properties := map[string]string{}
	if c.username != "" {
		properties["username"] = "set"
	}
	if c.authType != "" {
		properties["authtype"] = c.authType
	}
	if c.name != "" {
		properties["name"] = "set"
	}

	if c.passwordEncryption != "" {
		properties["passwordEncryption"] = "set"
	}
	telemetry.TrackEvent(eventName, properties)
	telemetry.CloseTelemetry()
}
