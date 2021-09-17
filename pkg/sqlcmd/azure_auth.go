// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"database/sql/driver"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	mssql "github.com/denisenkom/go-mssqldb"
)

const (
	ActiveDirectoryDefault     = "ActiveDirectoryDefault"
	ActiveDirectoryIntegrated  = "ActiveDirectoryIntegrated"
	ActiveDirectoryPassword    = "ActiveDirectoryPassword"
	ActiveDirectoryInteractive = "ActiveDirectoryInteractive"
	//ActiveDirectoryDeviceCodeFlow   = "ActiveDirectoryDeviceCodeFlow"
	ActiveDirectoryManagedIdentity  = "ActiveDirectoryManagedIdentity"
	ActiveDirectoryServicePrincipal = "ActiveDirectoryServicePrincipal"
	SqlPassword                     = "SqlPassword"
	NotSpecified                    = "NotSpecified"
	SqlClientId                     = "a94f9c62-97fe-4d19-b06d-416370FC77"
)

func azureTenantId() string {
	t := os.Getenv("AZURE_TENANT_ID")
	if t == "" {
		t = "common"
	}
	return t
}

func (s *Sqlcmd) GetTokenBasedConnection(connstr string, user string, password string) (driver.Connector, error) {
	var cred azcore.TokenCredential
	var err error
	t := azureTenantId()
	switch s.Connect.AuthenticationMethod {
	case ActiveDirectoryDefault:
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	case ActiveDirectoryInteractive:
		cred, err = azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{TenantID: t, ClientID: SqlClientId})
	case ActiveDirectoryPassword:
		cred, err = azidentity.NewUsernamePasswordCredential(t, SqlClientId, user, password, nil)
	case ActiveDirectoryManagedIdentity:
		cred, err = azidentity.NewManagedIdentityCredential(user, nil)
	case ActiveDirectoryServicePrincipal:
		cred, err = azidentity.NewClientSecretCredential(t, user, password, nil)
	default:
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	}

	if err != nil {
		return nil, err
	}

	conn, err := mssql.NewAccessTokenConnector(connstr, func() (string, error) {
		opts := policy.TokenRequestOptions{Scopes: []string{"https://database.windows.net/.default"}}
		tk, err := cred.GetToken(context.Background(), opts)
		if err != nil {
			return "", err
		}
		return tk.Token, err
	})

	return conn, err
}
