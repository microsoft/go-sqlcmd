// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// TestCurrentContextEndpointHasContainer verifies the function returns false when
// no current context
func TestCurrentContextEndpointHasContainer(t *testing.T) {
	SetFileName(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd", "sqlconfig-TestCurrentContextEndpointHasContainer"))
	Clean()

	assert.False(t, CurrentContextEndpointHasContainer())
}

func TestGetContainerId(t *testing.T) {
	SetFileName(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd", "sqlconfig-TestGetContainerId"))
	Clean()

	assert.Panics(t, func() { ContainerId() })
}

func TestGetContainerId2(t *testing.T) {
	SetFileName(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd", "sqlconfig-TestGetContainerId2"))
	Clean()

	AddEndpoint(Endpoint{
		AssetDetails: &AssetDetails{},
		EndpointDetails: EndpointDetails{
			Address: "127.0.0.1",
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
	assert.Panics(t, func() { ContainerId() })
}

func TestGetContainerId3(t *testing.T) {
	SetFileName(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd", "sqlconfig-TestGetContainerId3"))
	Clean()

	AddEndpoint(Endpoint{
		AssetDetails: &AssetDetails{
			ContainerDetails: &ContainerDetails{
				Id:    strings.Repeat("9", 32),
				Image: "www.image.url"}},
		EndpointDetails: EndpointDetails{
			Address: "127.0.0.1",
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

	assert.Panics(t, func() { ContainerId() })
}

func TestGetContainerId4(t *testing.T) {
	SetFileName(pal.FilenameInUserHomeDotDirectory(
		".sqlcmd", "sqlconfig-TestGetContainerId4"))
	Clean()
	AddEndpoint(Endpoint{
		AssetDetails: &AssetDetails{
			ContainerDetails: &ContainerDetails{
				Id:    strings.Repeat("9", 32),
				Image: "www.image.url"}},
		EndpointDetails: EndpointDetails{
			Address: "127.0.0.1",
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
	config.CurrentContext = "badbad"
	assert.Panics(t, func() { ContainerId() })
}
