// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	. "github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"reflect"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
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
							Password:          "weak",
						},
					}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config = tt.args.Config
			Clean()
			IsEmpty()
			GetConfigFileUsed()
			GetRedactedConfig(false)
			GetRedactedConfig(true)

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

			EndpointsExists()
			EndpointExists("endpoint")
			GetEndpoint("endpoint")
			OutputEndpoints(output.Struct, true)
			OutputEndpoints(output.Struct, false)
			FindFreePortForTds()
			DeleteEndpoint("endpoint2")

			user := User{
				Name:               "user",
				AuthenticationType: "basic",
				BasicAuth: &BasicAuthDetails{
					Username:          "username",
					PasswordEncrypted: false,
					Password:          "password",
				},
			}

			AddUser(user)
			AddUser(user)
			UserExists("user")
			GetUser("user")
			UserNameExists("username")
			OutputUsers(output.Struct, true)
			OutputUsers(output.Struct, false)
			DeleteUser("user")
			DeleteUser("user2")

			addContext()
			GetContext("context")
			OutputContexts(output.Struct, true)
			OutputContexts(output.Struct, false)
			DeleteContext("context")
			addContext()
			addContext()
			SetCurrentContextName("context")
			GetContainerId()
			GetCurrentContext()
			RemoveCurrentContext()
			RemoveCurrentContext()
			AddContextWithContainer("context", "imageName", 1433, "containerId", "user", "password", false)
			RemoveCurrentContext()
			DeleteEndpoint("endpoint")
		})
	}
}

func addContext() {
	user := "user1"
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
			if gotExists := UserExists(tt.args.name); gotExists != tt.wantExists {
				t.Errorf("UserExists() = %v, want %v", gotExists, tt.wantExists)
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

			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
			}()

			AddContextWithContainer(tt.args.contextName, tt.args.imageName, tt.args.portNumber, tt.args.containerId, tt.args.username, tt.args.password, tt.args.encryptPassword)
		})
	}
}
