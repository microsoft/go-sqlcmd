// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlconfig

/*
Package sqlconfig defines the schema for the sqlconfig file. The sqlconfig file
by default resides in the folder:

	Windows: %USERPROFILE%\.sqlcmd\sqlconfig
	*nix: ~/.sqlcmd/sqlconfig

The sqlconfig contains Contexts.  A context is named (e.g. mssql2) and
contains the Endpoint details (to connect to) and User details (to
use for authentication with the endpoint.

If there is more than one context defined, there is always a "currentcontext",
the currentcontext can be changed using

	sqlcmd config use-context CONTEXT_NAME

# Example

An example of the sqlconfig file looks like this:

	apiversion: v1
	endpoints:
	- asset:
		- container:
			id: 0e698e65e19d9c
			image: mcr.microsoft.com/mssql/server:2022-latest
	  endpoint:
		address: 127.0.0.1
		port: 1435
	  name: mssql
	contexts:
	- context:
		endpoint: mssql
		user: your-alias@mssql
	  name: mssql
	currentcontext: mssql
	kind: Config
	users:
	- user:
		username: your-alias
		password: REDACTED
	  name: your-alias@mssql

# Security

  - OnWindows the password is encrypted using the DPAPI.
  - TODO: On MacOS the password will be encrypted using the KeyChain

The password is also base64 encoded.

To view the decrypted and (base64) decoded passwords run

	sqlcmd config view --raw
*/
