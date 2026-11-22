// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

import (
	"fmt"
	. "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestConnect(t *testing.T) {
	Initialize(func(err error) {
		if err != nil {
			panic(err)
		}
	}, func(format string, a ...any) { fmt.Printf(format, a...) }, secret.Decode)

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
						Username:           "foo",
						PasswordEncryption: "dpapi",
						Password:           "bar",
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
			mssql := New(SqlOptions{})

			// If test name ends in 'Panic' expect a Panic
			if strings.HasSuffix(tt.name, "Panic") {
				assert.Panics(t, func() {
					mssql.Connect(tt.args.endpoint, tt.args.user, ConnectOptions{})
					mssql.Query("SELECT @@version")
				})
			} else {
				mssql.Connect(tt.args.endpoint, tt.args.user, ConnectOptions{})
				mssql.Query("SELECT @@version")
			}
		})
	}
}

func TestQueryExitOnError(t *testing.T) {
	password := os.Getenv("SQLCMDPASSWORD")
	if password == "" {
		t.Skip("SQLCMDPASSWORD is not set")
	}
	server := os.Getenv("SQLCMDSERVER")
	if server == "" {
		server = "localhost"
	}
	port := 1433
	if p := os.Getenv("SQLCMDPORT"); p != "" {
		var err error
		port, err = strconv.Atoi(p)
		if err != nil {
			t.Fatalf("SQLCMDPORT %q is not a number", p)
		}
	}
	userName := os.Getenv("SQLCMDUSER")
	if userName == "" {
		userName = "sa"
	}

	var queryErr error
	Initialize(func(err error) {
		if err != nil {
			queryErr = err
		}
	}, func(string, ...any) {}, secret.Decode)

	connect := func(t *testing.T, exitOnError bool) Sql {
		queryErr = nil
		s := New(SqlOptions{})
		s.Connect(
			Endpoint{EndpointDetails: EndpointDetails{Address: server, Port: port}},
			&User{
				AuthenticationType: "basic",
				BasicAuth: &BasicAuthDetails{
					Username:           userName,
					PasswordEncryption: "none",
					Password:           secret.Encode(password, "none"),
				},
			},
			ConnectOptions{Database: "master", ExitOnError: exitOnError})
		if queryErr != nil {
			t.Fatalf("connect failed: %v", queryErr)
		}
		return s
	}

	t.Run("server error is reported and the next query succeeds", func(t *testing.T) {
		s := connect(t, true)

		queryErr = nil
		s.Query("RAISERROR(N'restore failed', 16, 1)")
		assert.Error(t, queryErr)

		queryErr = nil
		s.Query("PRINT N'ok'")
		assert.NoError(t, queryErr)
	})

	t.Run("ScalarString reports a server error", func(t *testing.T) {
		s := connect(t, true)

		queryErr = nil
		s.ScalarString("RAISERROR(N'restore failed', 16, 1)")
		assert.Error(t, queryErr)
	})

	t.Run("informational message is not an error", func(t *testing.T) {
		s := connect(t, true)

		queryErr = nil
		s.Query("RAISERROR(N'progress', 10, 1)")
		assert.NoError(t, queryErr)
	})

	t.Run("server error is only printed without ExitOnError", func(t *testing.T) {
		s := connect(t, false)

		queryErr = nil
		s.Query("RAISERROR(N'restore failed', 16, 1)")
		assert.NoError(t, queryErr)
	})
}
