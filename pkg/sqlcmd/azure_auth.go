// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"database/sql/driver"
	"net/url"
	"os"

	"github.com/denisenkom/go-mssqldb/azuread"
)

const (
	NotSpecified = "NotSpecified"
	SqlPassword  = "SqlPassword"
	sqlClientId  = "a94f9c62-97fe-4d19-b06d-472bed8d2bcf"
)

func getSqlClientId() string {
	if clientId := os.Getenv("SQLCMDCLIENTID"); clientId != "" {
		return clientId
	}
	return sqlClientId
}

func (s *Sqlcmd) GetTokenBasedConnection(connstr string, user string, password string) (driver.Connector, error) {

	connectionUrl, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	if user != "" {
		connectionUrl.User = url.UserPassword(user, password)
	}

	query := connectionUrl.Query()
	query.Set("fedauth", s.Connect.authenticationMethod())
	query.Set("applicationclientid", getSqlClientId())

	switch s.Connect.AuthenticationMethod {
	case azuread.ActiveDirectoryServicePrincipal:
	case azuread.ActiveDirectoryApplication:
		query.Set("clientcertpath", os.Getenv("AZURE_CLIENT_CERTIFICATE_PATH"))
	case azuread.ActiveDirectoryInteractive:
		// AAD interactive needs minutes at minimum
		if s.Connect.LoginTimeoutSeconds < 120 {
			query.Set("connection timeout", "120")
		}
	}

	connectionUrl.RawQuery = query.Encode()
	return azuread.NewConnector(connectionUrl.String())
}
