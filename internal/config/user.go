// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	. "github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
	"strconv"
)

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

func DeleteUser(name string) {
	if UserExists(name) {
		ordinal := userOrdinal(name)
		config.Users = append(config.Users[:ordinal], config.Users[ordinal+1:]...)
		Save()
	}
}

func UserNameExists(name string) (exists bool) {
	for _, v := range config.Users {
		if v.Name == name {
			exists = true
			break
		}
	}

	return
}

func UserExists(name string) (exists bool) {
	for _, v := range config.Users {
		if name == v.Name {
			exists = true
			break
		}
	}
	return
}

func userOrdinal(name string) (ordinal int) {
	for i, c := range config.Users {
		if name == c.Name {
			ordinal = i
			break
		}
	}
	return
}

func GetUser(name string) (user User) {
	for _, v := range config.Users {
		if name == v.Name {
			user = v
			break
		}
	}
	return
}

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
