// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"database/sql/driver"
	"fmt"
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

func GetTokenBasedConnection(connstr string, authenticationMethod string) (driver.Connector, error) {

	connectionUrl, err := url.Parse(connstr)
	if err != nil {
		return nil, err
	}

	query := connectionUrl.Query()
	query.Set("fedauth", authenticationMethod)
	query.Set("applicationclientid", getSqlClientId())
	switch authenticationMethod {
	case azuread.ActiveDirectoryServicePrincipal, azuread.ActiveDirectoryApplication:
		query.Set("clientcertpath", os.Getenv("AZURE_CLIENT_CERTIFICATE_PATH"))
	case azuread.ActiveDirectoryInteractive:
		loginTimeout := query.Get("connection timeout")
		loginTimeoutSeconds := 0
		if loginTimeout != "" {
			_, _ = fmt.Sscanf(loginTimeout, "%d", &loginTimeoutSeconds)
		}
		// AAD interactive needs minutes at minimum
		if loginTimeoutSeconds > 0 && loginTimeoutSeconds < 120 {
			query.Set("connection timeout", "120")
		}
	}

	connectionUrl.RawQuery = query.Encode()
	return azuread.NewConnector(connectionUrl.String())
}
