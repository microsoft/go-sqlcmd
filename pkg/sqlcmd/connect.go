// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/microsoft/go-mssqldb/azuread"
	"github.com/microsoft/go-mssqldb/msdsn"
)

// ConnectSettings specifies the settings for SQL connections and queries
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
	// ignore error
	IgnoreError bool
	// ErrorSeverityLevel sets the minimum SQL severity level to treat as an error
	ErrorSeverityLevel uint8
	// Database is the name of the database for the connection
	Database string
	// ApplicationName is the name of the application to be included in the connection string
	ApplicationName string
	// DedicatedAdminConnection forces the connection to occur over tcp on the dedicated admin port. Requires Browser service access
	DedicatedAdminConnection bool
	// EnableColumnEncryption enables support for transparent column encryption
	EnableColumnEncryption bool
	// ChangePassword is the new password for the user to set during login
	ChangePassword string
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

func (connect ConnectSettings) RequiresPassword() bool {
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
	serverName, instance, port, protocol, err := splitServer(connect.ServerName)
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

	if strings.HasPrefix(serverName, `\\`) {
		// passing a pipe name of the format \\server\pipe\<pipename>
		pipeParts := strings.SplitN(string(serverName[2:]), `\`, 3)
		if len(pipeParts) != 3 {
			return "", &InvalidServerName
		}
		serverName = pipeParts[0]
		query.Add(msdsn.Pipe, pipeParts[2])
	}
	if port > 0 {
		connectionURL.Host = fmt.Sprintf("%s:%d", serverName, port)
	} else {
		connectionURL.Host = serverName
	}
	if connect.Database != "" {
		query.Add(msdsn.Database, connect.Database)
	}

	if connect.TrustServerCertificate {
		query.Add(msdsn.TrustServerCertificate, "true")
	}
	if connect.ApplicationIntent != "" && connect.ApplicationIntent != "default" {
		query.Add(msdsn.ApplicationIntent, connect.ApplicationIntent)
	}
	if connect.LoginTimeoutSeconds > 0 {
		query.Add(msdsn.DialTimeout, fmt.Sprint(connect.LoginTimeoutSeconds))
	}
	if connect.PacketSize > 0 {
		query.Add(msdsn.PacketSize, fmt.Sprint(connect.PacketSize))
	}
	if connect.WorkstationName != "" {
		query.Add(msdsn.WorkstationID, connect.WorkstationName)
	}
	if connect.Encrypt != "" && connect.Encrypt != "default" {
		query.Add(msdsn.Encrypt, connect.Encrypt)
	}
	if connect.LogLevel > 0 {
		query.Add(msdsn.LogParam, fmt.Sprint(connect.LogLevel))
	}
	if protocol != "" {
		query.Add(msdsn.Protocol, protocol)
	}
	if connect.ApplicationName != "" {
		query.Add(msdsn.AppName, connect.ApplicationName)
	}
	if connect.DedicatedAdminConnection {
		query.Set(msdsn.Protocol, "admin")
	}
	if connect.EnableColumnEncryption {
		query.Set("columnencryption", "true")
	}
	if connect.ChangePassword != "" {
		query.Set(msdsn.ChangePassword, connect.ChangePassword)
	}
	connectionURL.RawQuery = query.Encode()
	return connectionURL.String(), nil
}
