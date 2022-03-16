# Go-based SQL Utilities - Preview 

This repo contains command line tools and go packages for working with Microsoft SQL Server, Azure SQL Database, and Azure Synapse.

## Sqlcmd

The `sqlcmd` project aims to be a complete port of the native sqlcmd to the `go` language, utilizing the [go-mssqldb](https://github.com/denisenkom/go-mssqldb) driver. For full documentation of the tool, see https://docs.microsoft.com/sql/tools/sqlcmd-utility

### Breaking changes

We will be implementing command line switches and behaviors over time. Several switches and behaviors are expected to change in this implementation.

- `-P` switch will be removed. Passwords for SQL authentication can only be provided through these mechanisms:

    - The `SQLCMDPASSWORD` environment variable
    - The `:CONNECT` command
    - When prompted, the user can type the password to complete a connection (pending [#50](https://github.com/microsoft/go-sqlcmd/issues/50))
- `-r` requires a 0 or 1 argument
- `-R` switch will be removed. The go runtime does not provide access to user locale information, and it's not readily available through syscall on all supported platforms.
- `-I` switch will be removed. To disable quoted identifier behavior, add `SET QUOTED IDENTIFIER OFF` in your scripts.
- `-N` now takes a string value that can be one of `true`, `false`, or `disable` to specify the encryption choice. (`default` is the same as omitting the parameter)
  - If `-N` and `-C` are not provided, sqlcmd will negotiate authentication with the server without validating the server certificate.
  - If `-N` is provided but `-C` is not, sqlcmd will require validation of the server certificate. Note that a `false` value for encryption could still lead to encryption of the login packet.
  - If both `-N` and `-C` are provided, sqlcmd will use their values for encryption negotiation.
  - More information about client/server encryption negotiation can be found at <https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-tds/60f56408-0188-4cd5-8b90-25c6f2423868>
- Some behaviors that were kept to maintain compatibility with `OSQL` may be changed, such as alignment of column headers for some data types.
- All commands must fit on one line, even `EXIT`. Interactive mode will not check for open parentheses or quotes for commands and prompt for successive lines. The ODBC sqlcmd allows the query run by `EXIT(query)` to span multiple lines.

### Miscellaneous enhancements

- `:Connect` now has an optional `-G` parameter to select one of the authentication methods for Azure SQL Database  - `SqlAuthentication`, `ActiveDirectoryDefault`, `ActiveDirectoryIntegrated`, `ActiveDirectoryServicePrincipal`, `ActiveDirectoryManagedIdentity`, `ActiveDirectoryPassword`. If `-G` is not provided, either Integrated security or SQL Authentication will be used, dependent on the presence of a `-U` user name parameter.
- The new `--driver-logging-level` command line parameter allows you to see traces from the `go-mssqldb` client driver. Use `64` to see all traces.
- Sqlcmd can now print results using a vertical format. Use the new `-F vertical` command line option to set it. It's also controlled by the `SQLCMDFORMAT` scripting variable.

### Azure Active Directory Authentication

This version of sqlcmd supports a broader range of AAD authentication models, based on the [azidentity package](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity). The implementation relies on an AAD Connector in the [driver](https://github.com/denisenkom/go-mssqldb).

#### Command line

To use AAD auth, you can use one of two command line switches

`-G` is (mostly) compatible with its usage in the prior version of sqlcmd. If a user name and password are provided, it will authenticate using AAD Password authentication. If a user name is provided it will use AAD Interactive authentication which may display a web browser. If no user name or password is provided, it will use a DefaultAzureCredential which attempts to authenticate through a variety of mechanisms.

`--authentication-method=` can be used to specify one of the following authentication types.

`ActiveDirectoryDefault`

- For an overview of the types of authentication this mode will use, see (<https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity#defaultazurecredential>).
- Choose this method if your database automation scripts are intended to run in both local development environments and in a production deployment in Azure. You'll be able to use a client secret or an Azure CLI login on your development environment and a managed identity or client secret on your production deployment without changing the script.
- Setting environment variables AZURE_TENANT_ID, and AZURE_CLIENT_ID are necessary for DefaultAzureCredential to begin checking the environment configuration and look for one of the following additional environment variables in order to authenticate:

  - Setting environment variable AZURE_CLIENT_SECRET configures the DefaultAzureCredential to choose ClientSecretCredential.
  - Setting environment variable AZURE_CLIENT_CERTIFICATE_PATH configures the DefaultAzureCredential to choose ClientCertificateCredential if AZURE_CLIENT_SECRET is not set.
  - Setting environment variable AZURE_USERNAME configures the DefaultAzureCredential to choose UsernamePasswordCredential if AZURE_CLIENT_SECRET and AZURE_CLIENT_CERTIFICATE_PATH are not set.

`ActiveDirectoryIntegrated`

This method is currently not implemented and will fall back to `ActiveDirectoryDefault`

`ActiveDirectoryPassword`

This method will authenticate using a user name and password. It will not work if MFA is required.
You provide the user name and password using the usual command line switches or SQLCMD environment variables.
Set `AZURE_TENANT_ID` environment variable to the tenant id of the server if not using the default tenant of the user.

`ActiveDirectoryInteractive`

This method will launch a web browser to authenticate the user.

`ActiveDirectoryManagedIdentity`

Use this method when running sqlcmd on an Azure VM that has either a system-assigned or user-assigned managed identity. If using a user-assigned managed identity, set the user name to the ID of the managed identity. If using a system-assigned identity, leave user name empty.

`ActiveDirectoryServicePrincipal`

This method authenticates the provided user name as a service principal id and the password as the client secret for the service principal. Provide a user name in the form `<service principal id>@<tenant id>`. Set `SQLCMDPASSWORD` variable to the client secret. If using a certificate instead of a client secret, set `AZURE_CLIENT_CERTIFICATE_PATH` environment variable to the path of the certificate file.

### Environment variables for AAD auth

Some settings for AAD auth do not have command line inputs, and some environment variables are consumed directly by the `azidentity` package used by `sqlcmd`.
These environment variables can be set to configure some aspects of AAD auth and to bypass default behaviors. In addition to the variables listed above, the following are sqlcmd-specific and apply to multiple methods.

`SQLCMDCLIENTID` - set this to the identifier of an application registered in your AAD which is authorized to authenticate to Azure SQL Database. Applies to `ActiveDirectoryInteractive` and `ActiveDirectoryPassword` methods.

### Packages

#### sqlcmd executable

Build [sqlcmd](cmd/sqlcmd)

```sh

go build ./cmd/sqlcmd

```

#### sqlcmd package

pkg/sqlcmd is consumable by other hosts. Go docs for the package are forthcoming. See the test code and [main.go](cmd/sqlcmd/main.go) for examples of initializing and running sqlcmd.

## Building

Scripts to build the binaries and package them for release will be added in a build folder off the root. We will also add Azure Devops pipeline yml files there to initiate builds and releases. Until then just use `go build ./cmd/sqlcmd` to create a sqlcmd binary.

## Testing

The tests rely on SQLCMD scripting variables to provide the connection string parameters. Set SQLCMDSERVER, SQLCMDDATABASE, SQLCMDUSER, SQLCMDPASSWORD variables appropriately then

```sh

go test ./...

```

If you are developing on Windows, you can use docker or WSL to run the tests on Linux. `docker run` lets you pass the environment variables. For example, if your code is in `i:\git\go-sqlcmd` you can run tests in a docker container:

```cmd

docker run -rm -e SQLCMDSERVER=<yourserver> -e SQLCMDUSER=<youruser> -e SQLCMDPASSWORD=<yourpassword> -v i:\git\go-sqlcmd:/go-sqlcmd -w /go-sqlcmd golang:1.16 go test ./...

```

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.

