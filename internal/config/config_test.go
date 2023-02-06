// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	o := output.New(output.Options{LoggingLevel: 4, ErrorHandler: errorCallback, HintHandler: func(hints []string) {

	}})

	type args struct {
		Config Sqlconfig
	}
	tests := []struct {
		name string
		args args
	}{
		{"config",
			args{
				Config: Sqlconfig{
					Users: []User{{
						Name:               "user1",
						AuthenticationType: "basic",
						BasicAuth: &BasicAuthDetails{
							Username:          "user",
							PasswordEncrypted: false,
							Password:          secret.Encode("weak", false),
						},
					}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config = tt.args.Config
			SetFileName(pal.FilenameInUserHomeDotDirectory(
				".sqlcmd", "sqlconfig-TestConfig"))
			Clean()
			IsEmpty()
			GetConfigFileUsed()

			AddEndpoint(Endpoint{
				AssetDetails: &AssetDetails{
					ContainerDetails: &ContainerDetails{
						Id:    strings.Repeat("9", 64),
						Image: "www.image.url"},
				},
				EndpointDetails: EndpointDetails{
					Address: "localhost",
					Port:    1433,
				},
				Name: "endpoint",
			})

			AddEndpoint(Endpoint{
				EndpointDetails: EndpointDetails{
					Address: "localhost",
					Port:    1434,
				},
				Name: "endpoint",
			})

			AddEndpoint(Endpoint{
				EndpointDetails: EndpointDetails{
					Address: "localhost",
					Port:    1435,
				},
				Name: "endpoint",
			})

			EndpointsExists()
			EndpointExists("endpoint")
			GetEndpoint("endpoint")
			OutputEndpoints(o.Struct, true)
			OutputEndpoints(o.Struct, false)
			FindFreePortForTds()
			DeleteEndpoint("endpoint2")
			DeleteEndpoint("endpoint3")

			user := User{
				Name:               "user",
				AuthenticationType: "basic",
				BasicAuth: &BasicAuthDetails{
					Username:          "username",
					PasswordEncrypted: false,
					Password:          secret.Encode("password", false),
				},
			}

			AddUser(user)
			AddUser(user)
			AddUser(user)
			UserNameExists("user")
			GetUser("user")
			UserNameExists("username")
			OutputUsers(o.Struct, true)
			OutputUsers(o.Struct, false)

			DeleteUser("user3")

			RedactedConfig(true)
			RedactedConfig(false)

			addContext()
			addContext()
			addContext()
			GetContext("context")
			OutputContexts(o.Struct, true)
			OutputContexts(o.Struct, false)
			DeleteContext("context3")
			DeleteContext("context2")
			DeleteContext("context")

			addContext()
			addContext()

			SetCurrentContextName("context")
			CurrentContext()

			CurrentContextEndpointHasContainer()
			ContainerId()
			RemoveCurrentContext()
			RemoveCurrentContext()
			AddContextWithContainer("context", "imageName", 1433, "containerId", "user", "password", false)
			RemoveCurrentContext()
			DeleteEndpoint("endpoint")
			DeleteContext("context")
			DeleteUser("user2")
		})
	}
}

func addContext() {
	user := "user"
	AddContext(Context{
		ContextDetails: ContextDetails{
			Endpoint: "endpoint",
			User:     &user,
		},
		Name: "context",
	})
}

func TestDeleteUser(t *testing.T) {
	type args struct {
		name string
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteUser(tt.args.name)
		})
	}
}

func TestFindUniqueUserName(t *testing.T) {
	type args struct {
		name string
	}
	var tests []struct {
		name               string
		args               args
		wantUniqueUserName string
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUniqueUserName := FindUniqueUserName(tt.args.name); gotUniqueUserName != tt.wantUniqueUserName {
				t.Errorf("FindUniqueUserName() = %v, want %v", gotUniqueUserName, tt.wantUniqueUserName)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	type args struct {
		name string
	}
	var tests []struct {
		name     string
		args     args
		wantUser User
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUser := GetUser(tt.args.name); !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("GetUser() = %v, want %v", gotUser, tt.wantUser)
			}
		})
	}
}

func TestOutputUsers(t *testing.T) {
	type args struct {
		formatter func(interface{}) []byte
		detailed  bool
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			OutputUsers(tt.args.formatter, tt.args.detailed)
		})
	}
}

func TestUserExists(t *testing.T) {
	type args struct {
		name string
	}
	var tests []struct {
		name       string
		args       args
		wantExists bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotExists := UserNameExists(tt.args.name); gotExists != tt.wantExists {
				t.Errorf("UserNameExists() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func TestUserNameExists(t *testing.T) {
	type args struct {
		name string
	}
	var tests []struct {
		name       string
		args       args
		wantExists bool
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotExists := UserNameExists(tt.args.name); gotExists != tt.wantExists {
				t.Errorf("UserNameExists() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func Test_userOrdinal(t *testing.T) {
	type args struct {
		name string
	}
	var tests []struct {
		name        string
		args        args
		wantOrdinal int
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOrdinal := userOrdinal(tt.args.name); gotOrdinal != tt.wantOrdinal {
				t.Errorf("userOrdinal() = %v, want %v", gotOrdinal, tt.wantOrdinal)
			}
		})
	}
}

func TestAddContextWithContainerPanic(t *testing.T) {
	type args struct {
		contextName     string
		imageName       string
		portNumber      int
		containerId     string
		username        string
		password        string
		encryptPassword bool
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "AddContextWithContainerDefensePanics",
			args: args{"", "image", 1433, "id", "user", "password", false}},
		{name: "AddContextWithContainerDefensePanics",
			args: args{"context", "", 1433, "id", "user", "password", false}},
		{name: "AddContextWithContainerDefensePanics",
			args: args{"context", "image", 1433, "", "user", "password", false}},
		{name: "AddContextWithContainerDefensePanics",
			args: args{"context", "image", 0, "id", "user", "password", false}},
		{name: "AddContextWithContainerDefensePanics",
			args: args{"context", "image", 1433, "id", "", "password", false}},
		{name: "AddContextWithContainerDefensePanics",
			args: args{"context", "image", 1433, "id", "user", "", false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				AddContextWithContainer(tt.args.contextName, tt.args.imageName, tt.args.portNumber, tt.args.containerId, tt.args.username, tt.args.password, tt.args.encryptPassword)
			})
		})
	}
}

func TestConfig_AddContextWithNoEndpoint(t *testing.T) {
	assert.Panics(t, func() {

		user := "user1"
		AddContext(Context{
			ContextDetails: ContextDetails{
				Endpoint: "badbad",
				User:     &user,
			},
			Name: "context",
		})
	})
}

func TestConfig_GetCurrentContextWithNoContexts(t *testing.T) {
	assert.Panics(t, func() {

		CurrentContext()
	})
}

func TestConfig_GetCurrentContextEndPointNotFoundPanic(t *testing.T) {
	assert.Panics(t, func() {

		AddEndpoint(Endpoint{
			AssetDetails: &AssetDetails{
				ContainerDetails: &ContainerDetails{
					Id:    strings.Repeat("9", 64),
					Image: "www.image.url"},
			},
			EndpointDetails: EndpointDetails{
				Address: "localhost",
				Port:    1433,
			},
			Name: "endpoint",
		})

		user := "user1"
		AddContext(Context{
			ContextDetails: ContextDetails{
				Endpoint: "endpoint",
				User:     &user,
			},
			Name: "context",
		})

		DeleteEndpoint("endpoint")

		SetCurrentContextName("context")
		CurrentContext()
	})
}

func TestConfig_DeleteContextThatDoesNotExist(t *testing.T) {
	assert.Panics(t, func() {

		contextOrdinal("does-not-exist")
	})
}

func TestNegConfig_SetFileName(t *testing.T) {
	assert.Panics(t, func() {

		SetFileName("")
	})
}

func TestNegConfig_SetCurrentContextName(t *testing.T) {
	assert.Panics(t, func() {

		SetCurrentContextName("does not exist")
	})
}

func TestNegConfig_SetFileNameForTest(t *testing.T) {
	SetFileNameForTest(t)
}

func TestNegConfig_DefaultFileName(t *testing.T) {
	DefaultFileName()
}

func TestNegConfig_GetContext(t *testing.T) {
	assert.Panics(t, func() {

		GetContext("doesnotexist")
	})
}
