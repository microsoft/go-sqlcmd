// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	. "github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
	"strconv"
)

func AddEndpoint(endpoint Endpoint) (actualEndpointName string) {
	endpoint.Name = FindUniqueEndpointName(endpoint.Name)
	config.Endpoints = append(config.Endpoints, endpoint)
	Save()

	return endpoint.Name
}

func DeleteEndpoint(name string) {
	if EndpointExists(name) {
		ordinal := endpointOrdinal(name)
		config.Endpoints = append(config.Endpoints[:ordinal], config.Endpoints[ordinal+1:]...)
		Save()
	}
}

func EndpointsExists() (exists bool) {
	if len(config.Endpoints) > 0 {
		exists = true
	}

	return
}

func EndpointNameExists(name string) (exists bool) {
	for _, v := range config.Endpoints {
		if v.Name == name {
			exists = true
			break
		}
	}

	return
}

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

func endpointOrdinal(name string) (ordinal int) {
	for i, c := range config.Endpoints {
		if name == c.Name {
			ordinal = i
			break
		}
	}
	return
}

func GetEndpoint(name string) (endpoint Endpoint) {
	for _, e := range config.Endpoints {
		if name == e.Name {
			endpoint = e
			break
		}
	}
	return
}

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
