// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"strings"
	"testing"
)

// TestCurrentContextEndpointHasContainer verifies the function panics when
// no current context
func TestCurrentContextEndpointHasContainer(t *testing.T) {
	Clean()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	CurrentContextEndpointHasContainer()
}

func TestGetContainerId(t *testing.T) {
	Clean()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	ContainerId()
}

func TestGetContainerId2(t *testing.T) {
	Clean()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	AddEndpoint(Endpoint{
		AssetDetails: &AssetDetails{},
		EndpointDetails: EndpointDetails{
			Address: "localhost",
			Port:    1433,
		},
		Name: "endpoint",
	})

	user := "user"
	AddContext(Context{
		ContextDetails: ContextDetails{
			Endpoint: "endpoint",
			User:     &user,
		},
		Name: "context",
	})

	SetCurrentContextName("context")
	ContainerId()
}

func TestGetContainerId3(t *testing.T) {
	Clean()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	AddEndpoint(Endpoint{
		AssetDetails: &AssetDetails{
			ContainerDetails: &ContainerDetails{
				Id:    strings.Repeat("9", 32),
				Image: "www.image.url"}},
		EndpointDetails: EndpointDetails{
			Address: "localhost",
			Port:    1433,
		},
		Name: "endpoint",
	})

	user := "user"
	AddContext(Context{
		ContextDetails: ContextDetails{
			Endpoint: "endpoint",
			User:     &user,
		},
		Name: "context",
	})

	SetCurrentContextName("context")
	ContainerId()
}
