// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"net/url"

	"github.com/denisenkom/go-mssqldb/azuread"
)

// ConnectSettings specifies the settings for connections
type ConnectSettings struct {
	// ServerName is the full name including instance and port
	ServerName string
	// UseTrustedConnection indicates integrated auth is used when no user name is provided
	UseTrustedConnection bool
	// TrustServerCertificate sets the TrustServerCertificate setting on the connection string
	TrustServerCertificate bool
	// AuthenticationMethod defines the authentication method for connecting to Azure SQL Database
	AuthenticationMethod string
	// DisableEnvironmentVariables determines if sqlcmd resolves scripting variables from the process environment
	DisableEnvironmentVariables bool
	// DisableVariableSubstitution determines if scripting variables should be evaluated
	DisableVariableSubstitution bool
	// UserName is the username for the SQL connection
	UserName string
	// Password is the password used with SQL authentication or AAD authentications that require a password
	Password string
	// Encrypt is the choice of encryption
	Encrypt string
	// PacketSize is the size of the packet for TDS communication
	PacketSize int
	// LoginTimeoutSeconds specifies the timeout for establishing a connection
	LoginTimeoutSeconds int
	// WorkstationName is the string to use to identify the host in server DMVs
	WorkstationName string
	// ApplicationIntent can only be empty or "ReadOnly"
	ApplicationIntent string
	// LogLevel is the mssql driver log level
	LogLevel int
	// ExitOnError specifies whether to exit the app on an error
	ExitOnError bool
	// ErrorSeverityLevel sets the minimum SQL severity level to treat as an error
	ErrorSeverityLevel uint8
	// Database is the name of the database for the connection
	Database string
}

func (c ConnectSettings) authenticationMethod() string {
	if c.AuthenticationMethod == "" {
		return NotSpecified
	}
	return c.AuthenticationMethod
}

func (connect ConnectSettings) integratedAuthentication() bool {
	return connect.UseTrustedConnection || (connect.UserName == "" && connect.authenticationMethod() == NotSpecified)
}

func (connect ConnectSettings) sqlAuthentication() bool {
	return connect.authenticationMethod() == SqlPassword ||
		(!connect.UseTrustedConnection && connect.authenticationMethod() == NotSpecified && connect.UserName != "")
}

func (connect ConnectSettings) requiresPassword() bool {
	requiresPassword := connect.sqlAuthentication()
	if !requiresPassword {
		switch connect.authenticationMethod() {
		case azuread.ActiveDirectoryApplication, azuread.ActiveDirectoryPassword, azuread.ActiveDirectoryServicePrincipal:
			requiresPassword = true
		}
	}
	return requiresPassword
}

// ConnectionString returns the go-mssql connection string to use for queries
func (connect ConnectSettings) ConnectionString() (connectionString string, err error) {
	serverName, instance, port, err := splitServer(connect.ServerName)
	if serverName == "" {
		serverName = "."
	}
	if err != nil {
		return "", err
	}
	query := url.Values{}
	connectionURL := &url.URL{
		Scheme: "sqlserver",
		Path:   instance,
	}

	if connect.sqlAuthentication() || connect.authenticationMethod() == azuread.ActiveDirectoryPassword || connect.authenticationMethod() == azuread.ActiveDirectoryServicePrincipal || connect.authenticationMethod() == azuread.ActiveDirectoryApplication {
		connectionURL.User = url.UserPassword(connect.UserName, connect.Password)
	}
	if (connect.authenticationMethod() == azuread.ActiveDirectoryMSI || connect.authenticationMethod() == azuread.ActiveDirectoryManagedIdentity) && connect.UserName != "" {
		connectionURL.User = url.UserPassword(connect.UserName, connect.Password)
	}
	if port > 0 {
		connectionURL.Host = fmt.Sprintf("%s:%d", serverName, port)
	} else {
		connectionURL.Host = serverName
	}
	if connect.Database != "" {
		query.Add("database", connect.Database)
	}

	if connect.TrustServerCertificate {
		query.Add("trustservercertificate", "true")
	}
	if connect.ApplicationIntent != "" && connect.ApplicationIntent != "default" {
		query.Add("applicationintent", connect.ApplicationIntent)
	}
	if connect.LoginTimeoutSeconds > 0 {
		query.Add("connection timeout", fmt.Sprint(connect.LoginTimeoutSeconds))
	}
	if connect.PacketSize > 0 {
		query.Add("packet size", fmt.Sprint(connect.PacketSize))
	}
	if connect.WorkstationName != "" {
		query.Add("workstation id", connect.WorkstationName)
	}
	if connect.Encrypt != "" && connect.Encrypt != "default" {
		query.Add("encrypt", connect.Encrypt)
	}
	if connect.LogLevel > 0 {
		query.Add("log", fmt.Sprint(connect.LogLevel))
	}
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String(), nil
}
