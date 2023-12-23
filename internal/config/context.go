// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"errors"
	"fmt"
	"strconv"

	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
)

// AddContext adds the context to the sqlconfig file, and returns the context
// name, which maybe uniquified, if the passed in name already exists.
//
// Before calling this method, verify the Endpoint exists and give the user
// a descriptive error, (this function will panic, which should never be hit)
func AddContext(context Context) string {
	if !EndpointExists(context.Endpoint) {
		panic("Endpoint doesn't exist")
	}
	username := ""
	if context.User != nil {
		username = *context.User
	}
	context.Name = FindUniqueContextName(context.Name, username)

	// The command line parser sets user by default to "", in the case there
	// is no user (therefore Windows Authentication), we omit the
	// user from the sqlconfig file yaml by setting it to nil here.
	if context.User != nil && *context.User == "" {
		context.User = nil
	}

	config.Contexts = append(config.Contexts, context)
	Save()

	return context.Name
}

func AddAddOn(
	contextName, addOnName string,
	ContainerId string,
	Image string,
	address string,
	port int,
) {

	containerDetails := ContainerDetails{
		Id:    ContainerId,
		Image: Image}

	assetDetails := AssetDetails{
		ContainerDetails: &containerDetails}

	endpoint := Endpoint{
		AssetDetails: &assetDetails,
		EndpointDetails: EndpointDetails{
			Address: address,
			Port:    port},
		Name: addOnName + "@" + contextName,
	}

	uniqueEndpointName := AddEndpoint(endpoint)

	for i, c := range config.Contexts {
		if contextName == c.Name {
			config.Contexts[i].AddOns = append(config.Contexts[i].AddOns, AddOn{
				AddOnsDetails: AddOnsDetails{
					Type:     addOnName,
					Endpoint: uniqueEndpointName}})
			break
		}
	}

	Save()
}

// CurrentContextName returns the name of the current context in the configuration.
// The current context is the one that is currently active and used by the application.
func CurrentContextName() string {
	return config.CurrentContext
}

// ContextExists returns whether a context with the given name exists in the configuration.
// This function iterates over the list of contexts in the configuration and returns
// true if a context with the given name is found. Otherwise, the function returns false.
func ContextExists(name string) (exists bool) {
	for _, c := range config.Contexts {
		if name == c.Name {
			exists = true
			break
		}
	}
	return
}

// CurrentContext returns the current context's endpoint and user from the configuration.
// The function iterates over the list of contexts and endpoints in the configuration and returns the endpoint and user for the current context.
// If the current context does not have an endpoint, the function panics.
func CurrentContext() (endpoint Endpoint, user *User) {
	currentContextName := GetCurrentContextOrFatal()

	endPointFound := false
	for _, c := range config.Contexts {
		if c.Name == currentContextName {
			for _, e := range config.Endpoints {
				if e.Name == c.Endpoint {
					endpoint = e
					endPointFound = true
					break
				}
			}

			if UserExists(c) {
				for _, u := range config.Users {
					if u.Name == *c.User {
						user = &u
						break
					}
				}
			}
		}
	}

	if !endPointFound {
		panic(fmt.Sprintf(
			"Context '%v' has no endpoint.  Every context must have an endpoint",
			currentContextName,
		))
	}

	return
}

func CurrentContextAddOns() (addOns []AddOn) {
	currentContextName := GetCurrentContextOrFatal()

	for _, c := range config.Contexts {
		if c.Name == currentContextName {
			addOns = c.AddOns
			break
		}
	}

	return
}

// GetCurrentContextInfo returns endpoint and basic auth info
// associated with current context
func GetCurrentContextInfo() (server string, username string, password string) {
	endpoint, user := CurrentContext()
	server = fmt.Sprintf("%s,%d", endpoint.Address, endpoint.Port)
	if user != nil && user.AuthenticationType == "basic" {
		username = user.BasicAuth.Username
		if user.AuthenticationType == "basic" {
			password = decryptCallback(
				user.BasicAuth.Password,
				user.BasicAuth.PasswordEncryption,
			)
		}
	}
	return
}

// DeleteContext removes the context with the given name from the application's
// configuration. If the context does not exist, the function does nothing. The
// function also updates the CurrentContext field in the configuration to the
// first remaining context, or an empty string if no contexts remain. The
// updated configuration is saved to file.
func DeleteContext(name string) {
	if ContextExists(name) {
		ordinal := contextOrdinal(name)
		config.Contexts = append(config.Contexts[:ordinal], config.Contexts[ordinal+1:]...)

		if len(config.Contexts) > 0 {
			config.CurrentContext = config.Contexts[0].Name
		} else {
			config.CurrentContext = ""
		}

		Save()
	}
}

// FindUniqueContextName finds a unique context name, that is both a
// unique context name, but also a unique sa@context name.  If the name passed
// in is unique then this is returned, else we look for the name with a numeral
// postfix, starting at 2
func FindUniqueContextName(name string, username string) (uniqueContextName string) {
	if !ContextExists(name) &&
		!UserNameExists(username+"@"+name) {
		uniqueContextName = name
	} else {
		var postfixNumber = 2
		for {
			uniqueContextName = fmt.Sprintf(
				"%v%v",
				name,
				strconv.Itoa(postfixNumber),
			)
			if !ContextExists(uniqueContextName) {
				if !UserNameExists(username + "@" + uniqueContextName) {
					break
				}
			} else {
				postfixNumber++
			}
		}
	}

	return
}

// GetCurrentContextOrFatal returns the name of the current context in the
// configuration or panics if it is not set.
// This function first calls the CurrentContextName function to retrieve the
// current context's name, if the current context's name is empty, the function
// panics with an error message indicating that a context must be set.
// Otherwise, the current context's name is returned.
func GetCurrentContextOrFatal() (currentContextName string) {
	currentContextName = CurrentContextName()
	if currentContextName == "" {
		checkErr(errors.New(
			"no current context. To create a context use `sqlcmd create`, " +
				"e.g. `sqlcmd create mssql`"))
	}
	return
}

// SetCurrentContextName sets the current context in the configuration to the given name.
// If a context with the given name does not exist, the function panics.
// Otherwise, the CurrentContext field in the configuration object is updated
// with the given name and the configuration is saved to the file.
func SetCurrentContextName(name string) {
	if ContextExists(name) {
		config.CurrentContext = name
		Save()
	} else {
		panic("Context must exist")
	}
}

// RemoveCurrentContext removes the current context from the configuration.
// This function iterates over the list of contexts, endpoints, and users in the
// configuration and removes the current context, its endpoint, and its user.
// If there are no remaining contexts in the configuration after removing the
// current context, the CurrentContext field in the configuration object is set
// to an empty string. Otherwise, the CurrentContext field is set to the name
// of the first remaining context.
func RemoveCurrentContext() {
	currentContextName := config.CurrentContext

	for ci, c := range config.Contexts {
		if c.Name == currentContextName {
			for ei, e := range config.Endpoints {
				if e.Name == c.Endpoint {
					config.Endpoints = append(
						config.Endpoints[:ei],
						config.Endpoints[ei+1:]...)
					break
				}
			}

			for ui, u := range config.Users {
				if c.User != nil && u.Name == *c.User {
					config.Users = append(
						config.Users[:ui],
						config.Users[ui+1:]...)
					break
				}
			}

			config.Contexts = append(
				config.Contexts[:ci],
				config.Contexts[ci+1:]...)
			break
		}
	}

	if len(config.Contexts) > 0 {
		config.CurrentContext = config.Contexts[0].Name
	} else {
		config.CurrentContext = ""
	}
}

// GetContext retrieves a context from the configuration by its name.
// If the context does not exist, the function panics.
// If the context is not found, the function panics to indicate that the context must exist.
func GetContext(name string) (context Context) {
	for _, c := range config.Contexts {
		if name == c.Name {
			context = c
			return
		}
	}
	panic("Context does not exist")
}

// OutputContexts outputs the list of contexts in the configuration.
// The output can be either detailed, which includes all information about each context, or a list of context names only.
// This is controlled by the detailed flag, which is passed to the function.
func OutputContexts(formatter func(interface{}) []byte, detailed bool) {
	if detailed {
		formatter(config.Contexts)
	} else {
		var names []string

		for _, v := range config.Contexts {
			names = append(names, v.Name)
		}

		formatter(names)
	}
}

func contextOrdinal(name string) (ordinal int) {
	for i, c := range config.Contexts {
		if name == c.Name {
			ordinal = i
			return
		}
	}
	panic("Context not found")
}
