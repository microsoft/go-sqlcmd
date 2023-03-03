// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/io/folder"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"os"
	"path/filepath"
	"testing"
)

var config Sqlconfig
var filename string

// SetFileName sets the filename for the file that the application reads from and
// writes to. The file is created if it does not already exist, and Viper is configured
// to use the given filename.
func SetFileName(name string) {
	if name == "" {
		panic("name is empty")
	}

	filename = name

	file.CreateEmptyIfNotExists(filename)
	configureViper(filename)
}

func SetFileNameForTest(t *testing.T) {
	SetFileName(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd", "sqlconfig-"+t.Name()))
}

// DefaultFileName returns the default filename for the file that the application
// reads from and writes to. This is typically located in the user's home directory
// under the ".sqlcmd" directory. If an error occurs while attempting to retrieve
// the user's home directory, the function will return an empty string.
func DefaultFileName() (filename string) {
	home, err := os.UserHomeDir()
	if err != nil {
		trace(
			"Error getting user's home directory: %v, will use current directory %q as default",
			err,
			folder.Getwd(),
		)
		home = "."
	}
	filename = filepath.Join(home, ".sqlcmd", "sqlconfig")

	return
}

// Clean resets the application's configuration by setting the Users, Contexts,
// and Endpoints fields to nil, the CurrentContext field to an empty string,
// and saving the updated configuration. This effectively resets the configuration
// to its initial state.
func Clean() {
	config.Users = nil
	config.Contexts = nil
	config.Endpoints = nil
	config.CurrentContext = ""

	Save()
}

// IsEmpty returns a boolean indicating whether the application's configuration
// is empty. The configuration is considered empty if all of the following fields
// are empty or zero-valued: Users, Contexts, Endpoints, and CurrentContext.
// This function can be used to determine whether the configuration has been
// initialized or reset.
func IsEmpty() (isEmpty bool) {
	if len(config.Users) == 0 &&
		len(config.Contexts) == 0 &&
		len(config.Endpoints) == 0 &&
		config.CurrentContext == "" {
		isEmpty = true
	}

	return
}

// AddContextWithContainer adds a new context to the application's configuration
// with the given parameters. The context is associated with a container
// identified by its container ID. If any of the required parameters (i.e. containerId,
// imageName, portNumber, username, password, contextName) are empty or
// zero-valued, the function will panic. The function also ensures that the given
// contextName and username are unique, and it encrypts the password if
// requested. The updated configuration is saved to file.
func AddContextWithContainer(
	contextName string,
	imageName string,
	portNumber int,
	containerId string,
	username string,
	password string,
	encryptPassword bool,
) {
	if containerId == "" {
		panic("containerId must be provided")
	}
	if imageName == "" {
		panic("imageName must be provided")
	}
	if portNumber == 0 {
		panic("portNumber must be non-zero")
	}
	if username == "" {
		panic("username must be provided")
	}
	if password == "" {
		panic("password must be provided")
	}
	if contextName == "" {
		panic("contextName must be provided")
	}

	contextName = FindUniqueContextName(contextName, username)
	endPointName := FindUniqueEndpointName(contextName)
	userName := username + "@" + contextName

	config.CurrentContext = contextName

	config.Endpoints = append(config.Endpoints, Endpoint{
		AssetDetails: &AssetDetails{
			ContainerDetails: &ContainerDetails{
				Id:    containerId,
				Image: imageName},
		},
		EndpointDetails: EndpointDetails{
			Address: "127.0.0.1",
			Port:    portNumber,
		},
		Name: endPointName,
	})

	config.Contexts = append(config.Contexts, Context{
		ContextDetails: ContextDetails{
			Endpoint: endPointName,
			User:     &userName,
		},
		Name: contextName,
	})

	user := User{
		AuthenticationType: "basic",
		BasicAuth: &BasicAuthDetails{
			Username:          username,
			PasswordEncrypted: encryptPassword,
			Password:          encryptCallback(password, encryptPassword),
		},
		Name: userName,
	}

	config.Users = append(config.Users, user)

	Save()
}

// RedactedConfig function returns a Sqlconfig struct with the Users field
// having their BasicAuth password field either replaced with the decrypted
// password or the string "REDACTED", depending on the value of the raw
// parameter. This allows the caller to either get the full password or a
// redacted version, where the password is hidden.
func RedactedConfig(raw bool) (c Sqlconfig) {
	c = config
	for i := range c.Users {
		user := c.Users[i]
		if user.AuthenticationType == "basic" {
			if raw {
				user.BasicAuth.Password = decryptCallback(
					user.BasicAuth.Password,
					user.BasicAuth.PasswordEncrypted,
				)
			} else {
				user.BasicAuth.Password = "REDACTED"
			}
		}
	}

	return
}
