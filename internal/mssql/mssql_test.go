// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package mssql

import (
	. "github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"runtime"
	"strings"
	"testing"
)

func TestConnect(t *testing.T) {
	t.Skip() // BUG(stuartpa): Re-enable before merge
	endpoint := Endpoint{
		EndpointDetails: EndpointDetails{
			Address: "localhost",
			Port:    1433,
		},
		Name: "local-default-instance"}
	type args struct {
		endpoint Endpoint
		user     *User
		console  sqlcmd.Console
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "connectBasicPanic", args: args{
			endpoint: endpoint,
			user: &User{
				Name:               "basicUser",
				AuthenticationType: "basic",
				BasicAuth: &BasicAuthDetails{
					Username:          "foo",
					PasswordEncrypted: true,
					Password:          "bar",
				},
			},
			console: nil,
		},
			want: 0,
		},
		{
			name: "invalidAuthTypePanic", args: args{
			endpoint: endpoint,
			user: &User{
				Name:               "basicUser",
				AuthenticationType: "badbad",
			},
			console: nil,
		},
			want: 0,
		},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name string
			args args
			want int
		}{
			name: "connectTrusted", args: args{endpoint: endpoint, user: nil, console: nil},
			want: 0})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// If test name ends in 'Panic' expect a Panic
			if strings.HasSuffix(tt.name, "Panic") {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
			}

			s := Connect(tt.args.endpoint, tt.args.user, tt.args.console)
			Query(s, "SELECT @@version")
		})
	}
}
