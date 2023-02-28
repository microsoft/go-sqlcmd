// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"strconv"

	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
)

// AddUser adds a new user to the configuration.
// The user's name is first modified to be unique by calling the FindUniqueUserName function.
// If the user's authentication type is "basic", the user's BasicAuth field must be non-nil and the username must be non-empty.
// The new user is then added to the list of users in the configuration object and the configuration is saved to the file.
func AddUser(user User) {
	user.Name = FindUniqueUserName(user.Name)

	if user.AuthenticationType == "basic" {
		if user.BasicAuth == nil {
			panic("If authType is basic, then user.BasicAuth must be provided")
		}

		if user.BasicAuth.Username == "" {
			panic("BasicAuth Username cannot be empty")
		}
	}

	config.Users = append(config.Users, user)
	Save()
}

// DeleteUser removes a user from the configuration by their name.
// If the user does not exist, the function does nothing.
// Otherwise, the user is removed from the list of users in the configuration object and the configuration is saved to the file.
func DeleteUser(name string) {
	if UserNameExists(name) {
		ordinal := userOrdinal(name)
		config.Users = append(config.Users[:ordinal], config.Users[ordinal+1:]...)
		Save()
	}
}

// FindUniqueUserName generates a unique user name based on the given name.
// If the given name is not already in use, it is returned as-is.
// Otherwise, a number is appended to the end of the given name to make it unique.
// This number starts at 2 and is incremented until a unique user name is found.
func FindUniqueUserName(name string) (uniqueUserName string) {
	if !UserNameExists(name) {
		uniqueUserName = name
	} else {
		var postfixNumber = 2

		for {
			uniqueUserName = fmt.Sprintf(
				"%v%v",
				name,
				strconv.Itoa(postfixNumber),
			)
			if !UserNameExists(uniqueUserName) {
				break
			} else {
				postfixNumber++
			}
		}
	}

	return
}

// GetUser retrieves a user from the configuration by their name.
func GetUser(name string) (user User) {
	for _, v := range config.Users {
		if name == v.Name {
			user = v
			return
		}
	}
	panic("User must exist")
}

// OutputUsers outputs the list of users in the configuration.
// The output can be either detailed, which includes all information about each user, or a list of user names only.
// This is controlled by the detailed flag, which is passed to the function.
func OutputUsers(formatter func(interface{}) []byte, detailed bool) {
	if detailed {
		formatter(config.Users)
	} else {
		var names []string

		for _, v := range config.Users {
			names = append(names, v.Name)
		}

		formatter(names)
	}
}

// UserNameExists checks if a user with the given name exists in the configuration.
// It iterates over the list of users in the configuration object and returns true if a user with the given name is found.
// Otherwise, it returns false.
// This function can be useful for checking if a user with a given name already exists before adding a new user or updating an existing user.
func UserNameExists(name string) (exists bool) {
	for _, v := range config.Users {
		if v.Name == name {
			exists = true
			break
		}
	}

	return
}

// userOrdinal returns the index of a user in the list of users in the configuration object.
// If the user does not exist, the function returns -1.
// This function iterates over the list of users and returns the index of the user with the given name.
func userOrdinal(name string) (ordinal int) {
	for i, c := range config.Users {
		if name == c.Name {
			ordinal = i
			break
		}
	}
	return
}

// UserExists checks if the current context has a 'user', e.g. a context used
// for trusted authentication will not have a user.
func UserExists(context Context) bool {
	return context.ContextDetails.User != nil && *context.ContextDetails.User != ""
}
