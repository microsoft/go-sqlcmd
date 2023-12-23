// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import "fmt"

// This function gets the container ID of the current context's endpoint. It first
// checks if the current context exists and has an endpoint. Then it checks if the
// endpoint has a container and retrieves its ID. Otherwise, it returns the container ID.
func ContainerId() (containerId string) {
	currentContextName := config.CurrentContext

	if currentContextName == "" {
		panic("currentContextName must not be empty")
	}

	for _, c := range config.Contexts {
		if c.Name == currentContextName {
			for _, e := range config.Endpoints {
				if e.Name == c.Endpoint {
					if e.ContainerDetails == nil {
						panic("Endpoint does not have a container")
					}
					containerId = e.ContainerDetails.Id

					if len(containerId) != 64 {
						panic(fmt.Sprintf("container id must be 64 characters (id: %q)", containerId))
					}

					return
				}
			}
		}
	}
	panic("Id not found")
}

// CurrentContextEndpointHasContainer() checks if the current context endpoint
// has a container. If the endpoint has a AssetDetails.ContainerDetails field, the function
// returns true, otherwise it returns false.
func CurrentContextEndpointHasContainer() (exists bool) {
	currentContextName := config.CurrentContext

	if currentContextName == "" {
		panic("currentContextName must not be empty")
	}

	for _, c := range config.Contexts {
		if c.Name == currentContextName {
			for _, e := range config.Endpoints {
				if e.Name == c.Endpoint {
					if e.AssetDetails != nil {
						if e.AssetDetails.ContainerDetails != nil {
							exists = true
						}
					}
					break
				}
			}
		}
	}
	return
}

// FindFreePortForTds is used to find a free port number to use for the TDS
// protocol. It starts at port number 1433 and continues until it finds a port
// number that is not currently in use by any of the endpoints in the
// configuration. It also checks that the port is available on the local machine.
// If no available port is found after trying up to port number 5000, the function panics.
func FindFreePortForTds(startingPortNumber int) (portNumber int) {
	portNumber = startingPortNumber

	for {
		foundFreePortNumber := true
		for _, endpoint := range config.Endpoints {
			if endpoint.Port == portNumber {
				foundFreePortNumber = false
				break
			}
		}

		if foundFreePortNumber {
			// Check this port is actually available on the local machine
			if isLocalPortAvailableCallback(portNumber) {
				break
			}
		}

		portNumber++

		if portNumber == startingPortNumber+2000 {
			panic("Did not find an available port")
		}
	}

	return
}
