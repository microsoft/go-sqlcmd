// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/file"
	"os"
	"path/filepath"
)

var config Sqlconfig
var filename string

func SetFileName(name string) {
	if name == "" {
		panic("name is empty")
	}

	filename = name

	file.CreateEmptyIfNotExists(filename)
	configureViper(filename)
}

func DefaultFileName() (filename string) {
	home, err := os.UserHomeDir()
	checkErr(err)
	filename = filepath.Join(home, ".sqlcmd", "sqlconfig")

	return
}

func Clean() {
	config.Users = nil
	config.Contexts = nil
	config.Endpoints = nil
	config.CurrentContext = ""

	Save()
}

func IsEmpty() (isEmpty bool) {
	if len(config.Users) == 0 &&
		len(config.Contexts) == 0 &&
		len(config.Endpoints) == 0 &&
		config.CurrentContext == "" {
		isEmpty = true
	}

	return
}

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

	config.ApiVersion = "v1"
	config.Kind = "Config"
	config.CurrentContext = contextName

	config.Endpoints = append(config.Endpoints, Endpoint{
		AssetDetails: &AssetDetails{
			ContainerDetails: &ContainerDetails{
				Id:    containerId,
				Image: imageName},
		},
		EndpointDetails: EndpointDetails{
			Address: "localhost",
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

func GetRedactedConfig(raw bool) (c Sqlconfig) {
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
