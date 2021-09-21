// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"database/sql/driver"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	mssql "github.com/denisenkom/go-mssqldb"
)

const (
	ActiveDirectoryDefault          = "ActiveDirectoryDefault"
	ActiveDirectoryIntegrated       = "ActiveDirectoryIntegrated"
	ActiveDirectoryPassword         = "ActiveDirectoryPassword"
	ActiveDirectoryInteractive      = "ActiveDirectoryInteractive"
	ActiveDirectoryManagedIdentity  = "ActiveDirectoryManagedIdentity"
	ActiveDirectoryServicePrincipal = "ActiveDirectoryServicePrincipal"
	SqlPassword                     = "SqlPassword"
	NotSpecified                    = "NotSpecified"
	sqlClientId                     = "a94f9c62-97fe-4d19-b06d-472bed8d2bcf"
)

func azureTenantId() string {
	t := os.Getenv("AZURE_TENANT_ID")
	if t == "" {
		t = "common"
	}
	return t
}

var resourceMap = map[string]string{
	".database.chinacloudapi.cn":  "https://database.chinacloudapi.cn/",
	".database.cloudapi.de":       "https://database.cloudapi.de/",
	".database.usgovcloudapi.net": "https://database.usgovcloudapi.net/",
	".database.windows.net":       "https://database.windows.net/",
}

func (s *Sqlcmd) getResourceUrl() string {
	resource := os.Getenv("SQLCMDAZURERESOURCE")
	if resource == "" {
		server, _, _, _ := s.vars.SQLCmdServer()
		for k := range resourceMap {
			if strings.HasSuffix(strings.ToLower(server), k) {
				return resourceMap[k]
			}
		}
	}
	return "https://database.windows.net"
}

func getSqlClientId() string {
	if clientId := os.Getenv("SQLCMDCLIENTID"); clientId != "" {
		return clientId
	}
	return sqlClientId
}

func (s *Sqlcmd) GetTokenBasedConnection(connstr string, user string, password string) (driver.Connector, error) {
	var cred azcore.TokenCredential
	var err error
	scope := ".default"
	t := azureTenantId()
	switch s.Connect.AuthenticationMethod {
	case ActiveDirectoryDefault:
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	case ActiveDirectoryInteractive:
		cred, err = azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{TenantID: t, ClientID: getSqlClientId()})
		scope = "user_impersonation"
	case ActiveDirectoryPassword:
		cred, err = azidentity.NewUsernamePasswordCredential(t, getSqlClientId(), user, password, nil)
	case ActiveDirectoryManagedIdentity:
		cred, err = azidentity.NewManagedIdentityCredential(user, nil)
	case ActiveDirectoryServicePrincipal:
		cred, err = azidentity.NewClientSecretCredential(t, user, password, nil)
	default:
		// no implementation of AAD Integrated yet
		cred, err = azidentity.NewDefaultAzureCredential(nil)
	}

	if err != nil {
		return nil, err
	}
	resourceUrl := s.getResourceUrl()
	conn, err := mssql.NewAccessTokenConnector(connstr, func() (string, error) {
		opts := policy.TokenRequestOptions{Scopes: []string{resourceUrl + scope}}
		tk, err := cred.GetToken(context.Background(), opts)
		if err != nil {
			return "", err
		}
		return tk.Token, err
	})

	return conn, err
}
