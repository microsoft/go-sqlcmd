// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"strconv"
)

// AddEndpoint adds a new endpoint to the application's configuration with
// the given parameters. If the provided endpoint name is not unique, the
// function will modify it to ensure that it is unique before adding it to the
// configuration. The updated configuration is saved to file, and the function
// returns the actual endpoint name that was added. This may be different
// from the provided name if the original name was not unique.
func AddEndpoint(endpoint Endpoint) (actualEndpointName string) {
	endpoint.Name = FindUniqueEndpointName(endpoint.Name)
	config.Endpoints = append(config.Endpoints, endpoint)
	Save()

	return endpoint.Name
}

// DeleteEndpoint removes the endpoint with the given name from the application's
// configuration. If the endpoint does not exist, the function does nothing. The
// updated configuration is saved to file.
func DeleteEndpoint(name string) {
	if EndpointExists(name) {
		ordinal := endpointOrdinal(name)
		config.Endpoints = append(config.Endpoints[:ordinal], config.Endpoints[ordinal+1:]...)
		Save()
	}
}

// EndpointsExists returns whether there are any endpoints in the configuration.
// This function checks the length of the Endpoints field in the configuration
// object and returns true if it is greater than zero. Otherwise, the function returns false.
func EndpointsExists() (exists bool) {
	if len(config.Endpoints) > 0 {
		exists = true
	}

	return
}

// EndpointExists returns whether an endpoint with the given name exists in
// the configuration. This function iterates over the list of endpoints in the
// configuration and returns true if an endpoint with the given name is found.
// Otherwise, the function returns false.
func EndpointExists(name string) (exists bool) {
	if name == "" {
		panic("Name must not be empty")
	}

	for _, c := range config.Endpoints {
		if name == c.Name {
			exists = true
			break
		}
	}
	return
}

// EndpointNameExists returns whether an endpoint with the given name exists
// in the configuration. This function iterates over the list of endpoints in the
// configuration and returns true if an endpoint with the given name is found.
// Otherwise, the function returns false.
func EndpointNameExists(name string) (exists bool) {
	for _, v := range config.Endpoints {
		if v.Name == name {
			exists = true
			break
		}
	}

	return
}

// FindUniqueEndpointName returns a unique name for an endpoint with the
// given name.
// If an endpoint with the given name does not exist in the configuration, the
// function returns the given name. Otherwise, the function returns a modified
// version of the given name that includes a number at the end to make it unique.
func FindUniqueEndpointName(name string) (uniqueEndpointName string) {
	if !EndpointNameExists(name) {
		uniqueEndpointName = name
	} else {
		var postfixNumber = 2

		for {
			uniqueEndpointName = fmt.Sprintf(
				"%v%v",
				name,
				strconv.Itoa(postfixNumber),
			)
			if !EndpointNameExists(uniqueEndpointName) {
				break
			} else {
				postfixNumber++
			}
		}
	}

	return
}

// GetEndpoint returns the endpoint with the given name from the configuration.
func GetEndpoint(name string) (endpoint Endpoint) {
	for _, e := range config.Endpoints {
		if name == e.Name {
			endpoint = e
			break
		}
	}
	return
}

// OutputEndpoints outputs the list of endpoints in the configuration in a specified format.
// This function takes a formatter function and a flag indicating whether to
// output detailed information or just the names of the endpoints.
// If detailed information is requested, the formatter function is called with
// the list of endpoints in the configuration as the argument.
// Otherwise, the formatter function is called with a list of just the names of
// the endpoints in the configuration.
func OutputEndpoints(formatter func(interface{}) []byte, detailed bool) {
	if detailed {
		formatter(config.Endpoints)
	} else {
		var names []string

		for _, v := range config.Endpoints {
			names = append(names, v.Name)
		}

		formatter(names)
	}
}

func endpointOrdinal(name string) (ordinal int) {
	for i, c := range config.Endpoints {
		if name == c.Name {
			ordinal = i
			break
		}
	}
	return
}
