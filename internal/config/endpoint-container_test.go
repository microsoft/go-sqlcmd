package config

import (
	"strings"
	"testing"

	. "github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
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
	GetContainerId()
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
	GetContainerId()
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
	GetContainerId()
}

func TestGetContainerId4(t *testing.T) {
	Clean()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	SetCurrentContextName("badbad")

	GetContainerId()
}
