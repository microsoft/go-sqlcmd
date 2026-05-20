// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/stretchr/testify/assert"
)

func TestConnectionStringIncludesPasswordForAuthMethods(t *testing.T) {
	authMethodsRequiringPassword := []string{
		azuread.ActiveDirectoryPassword,
		azuread.ActiveDirectoryServicePrincipal,
		azuread.ActiveDirectoryApplication,
		azuread.ActiveDirectoryServicePrincipalAccessToken,
	}

	pwd := uuid.New().String()

	for _, method := range authMethodsRequiringPassword {
		t.Run(method, func(t *testing.T) {
			settings := ConnectSettings{
				ServerName:           "someserver",
				AuthenticationMethod: method,
				UserName:             "myapp@mytenant",
				Password:             pwd,
			}

			connectionString, err := settings.ConnectionString()
			if assert.NoError(t, err) {
				expected := fmt.Sprintf("sqlserver://myapp%%40mytenant:%s@someserver", pwd)
				assert.Equal(t, expected, connectionString,
					"auth method %q should include user:password in the connection URL", method)
			}
		})
	}
}

func TestConnectionStringExcludesPasswordForNonCredentialAuthMethods(t *testing.T) {
	authMethodsWithoutPassword := []string{
		azuread.ActiveDirectoryDefault,
		azuread.ActiveDirectoryIntegrated,
		azuread.ActiveDirectoryInteractive,
		azuread.ActiveDirectoryDeviceCode,
		azuread.ActiveDirectoryAzCli,
		azuread.ActiveDirectoryAzureDeveloperCli,
		azuread.ActiveDirectoryAzurePipelines,
		azuread.ActiveDirectoryEnvironment,
		azuread.ActiveDirectoryWorkloadIdentity,
		azuread.ActiveDirectoryClientAssertion,
		azuread.ActiveDirectoryOnBehalfOf,
	}

	pwd := uuid.New().String()

	for _, method := range authMethodsWithoutPassword {
		t.Run(method, func(t *testing.T) {
			settings := ConnectSettings{
				ServerName:           "someserver",
				AuthenticationMethod: method,
				UserName:             "myapp@mytenant",
				Password:             pwd,
			}

			connectionString, err := settings.ConnectionString()
			if assert.NoError(t, err) {
				assert.False(t, strings.Contains(connectionString, pwd),
					"auth method %q should not include password in the connection URL", method)
			}
		})
	}
}

func TestConnectionStringIncludesPasswordForManagedIdentityWithUserName(t *testing.T) {
	managedIdentityMethods := []string{
		azuread.ActiveDirectoryMSI,
		azuread.ActiveDirectoryManagedIdentity,
	}

	pwd := uuid.New().String()

	for _, method := range managedIdentityMethods {
		t.Run(method+"_with_username", func(t *testing.T) {
			settings := ConnectSettings{
				ServerName:           "someserver",
				AuthenticationMethod: method,
				UserName:             "myclientid",
				Password:             pwd,
			}

			connectionString, err := settings.ConnectionString()
			if assert.NoError(t, err) {
				expected := fmt.Sprintf("sqlserver://myclientid:%s@someserver", pwd)
				assert.Equal(t, expected, connectionString,
					"auth method %q with UserName should include user:password in the connection URL", method)
			}
		})

		t.Run(method+"_without_username", func(t *testing.T) {
			settings := ConnectSettings{
				ServerName:           "someserver",
				AuthenticationMethod: method,
				Password:             pwd,
			}

			connectionString, err := settings.ConnectionString()
			if assert.NoError(t, err) {
				assert.False(t, strings.Contains(connectionString, pwd),
					"auth method %q without UserName should not include password in the connection URL", method)
			}
		})
	}
}

func TestRequiresPassword(t *testing.T) {
	methodsThatRequirePassword := []string{
		azuread.ActiveDirectoryPassword,
		azuread.ActiveDirectoryServicePrincipal,
		azuread.ActiveDirectoryApplication,
		azuread.ActiveDirectoryServicePrincipalAccessToken,
	}

	for _, method := range methodsThatRequirePassword {
		t.Run(method+"_requires_password", func(t *testing.T) {
			settings := ConnectSettings{
				AuthenticationMethod: method,
				UserName:             "someuser",
			}
			assert.True(t, settings.RequiresPassword(),
				"auth method %q should require a password", method)
		})
	}

	methodsThatDontRequirePassword := []string{
		azuread.ActiveDirectoryDefault,
		azuread.ActiveDirectoryIntegrated,
		azuread.ActiveDirectoryInteractive,
		azuread.ActiveDirectoryDeviceCode,
		azuread.ActiveDirectoryAzCli,
		azuread.ActiveDirectoryAzureDeveloperCli,
		azuread.ActiveDirectoryAzurePipelines,
		azuread.ActiveDirectoryEnvironment,
		azuread.ActiveDirectoryWorkloadIdentity,
		azuread.ActiveDirectoryClientAssertion,
		azuread.ActiveDirectoryOnBehalfOf,
		azuread.ActiveDirectoryMSI,
		azuread.ActiveDirectoryManagedIdentity,
	}

	for _, method := range methodsThatDontRequirePassword {
		t.Run(method+"_does_not_require_password", func(t *testing.T) {
			settings := ConnectSettings{
				AuthenticationMethod: method,
				UserName:             "someuser",
			}
			assert.False(t, settings.RequiresPassword(),
				"auth method %q should not require a password", method)
		})
	}
}

func TestConnectionStringIncludesPasswordForSqlAuth(t *testing.T) {
	pwd := uuid.New().String()
	settings := ConnectSettings{
		ServerName: "someserver",
		UserName:   "someuser",
		Password:   pwd,
	}

	connectionString, err := settings.ConnectionString()
	if assert.NoError(t, err) {
		assert.True(t, strings.Contains(connectionString, pwd),
			"SQL authentication should include password in the connection URL")
	}
}
