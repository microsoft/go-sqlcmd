# SQLCMD CLI

This repo contains the `sqlcmd` command line tool and Go packages for working with Microsoft SQL Server, Azure SQL Database, and Azure Synapse.  

Learn more about how `sqlcmd` is used from a articles/posts written by the community: [Community Buzz][].

## Installation

`sqlcmd` is available in package managers for all major platforms.

### Windows

`sqlcmd` is available via [Winget][], [Choco][] and as a downloadable .msi or .zip from the [releases page][]. The .msi installer is signed with a Microsoft Authenticode certificate.

#### WinGet

| Install:                | Upgrade:                    |
| ----------------------- |-----------------------------|
| `winget install sqlcmd` | `winget upgrade sqlcmd`     |

#### Choco

| Install:                | Upgrade:                |
| ----------------------- |-------------------------|
| `choco install sqlcmd`  | `choco upgrade sqlcmd`  |

### macOS

`sqlcmd` is available via [Homebrew][], and as a downloadable .tar from the [releases page][].

#### Homebrew

| Install:              | Upgrade:              |
| --------------------- | --------------------- |
| `brew install sqlcmd` | `brew upgrade sqlcmd` |

##### Apple Silicon Macs (M1/M2)
Macs running Apple Silicon require [Docker][] Desktop to use Rosetta for x86/amd64 emulation. Follow these steps before creating a SQL Server instance:
- Open Docker Desktop.
- Go to the settings/preferences menu.
- Find the “Features in development” section.
- Enable the "Use Rosetta for x86/amd64 emulation on Apple Silicon" checkbox.

### Linux

`sqlcmd` is available via [Linuxbrew][], and as a downloadable .rpm/.deb and .tar from the [releases page][].

On Linux, `sqlcmd` is also available through `apt-get`, `yum` and `zypper` package managers.  Instructions can be found [here](https://learn.microsoft.com/sql/tools/sqlcmd/go-sqlcmd-utility?view=sql-server-ver16#download-and-install-go-sqlcmd).

#### Linuxbrew

The Homebrew package manager may be used on Linux and Windows Subsystem for Linux (WSL) 2. Homebrew was formerly referred to as Linuxbrew when running on Linux or WSL.

| Install:              | Upgrade:              |
| --------------------- | --------------------- |
| `brew install sqlcmd` | `brew upgrade sqlcmd` |

## Use sqlcmd to create local SQL Server and Azure SQL Edge instances

Use `sqlcmd` to create SQL Server and Azure SQL Edge instances using a local container runtime (e.g. [Docker][] or [Podman][])

### Create SQL Server instance using local container runtime and connect using Azure Data Studio

To create a local SQL Server instance with the AdventureWorksLT database restored, query it, and connect to it using Azure Data Studio, run:

```
sqlcmd create mssql --accept-eula --using https://aka.ms/AdventureWorksLT.bak
sqlcmd query "SELECT DB_NAME()"
sqlcmd open ads
```

Use `sqlcmd --help` to view all the available sub-commands.  Use `sqlcmd -?` to view the original ODBC `sqlcmd` flags.

### The ~/.sqlcmd/sqlconfig file

Each time `sqlcmd create` completes, a new context is created (e.g. mssql, mssql2, mssql3 etc.).  A context contains the endpoint and user configuration detail.  To switch between contexts, run `sqlcmd config use <context-name>`, to view name of the current context, run `sqlcmd config current-context`, to list all contexts, run `sqlcmd config get-contexts`.

To view connection strings (ODBC/ADO.NET/JDBC etc.) for the current context and user & endpoint details for all contexts held in the `~/.sqlcmd/sqlconfig` file:

```
sqlcmd config connection-strings
sqlcmd config view
```

### Versions

To see all version tags to choose from (2017, 2019, 2022 etc.), and install a specific version, run:

```
SET SQLCMD_ACCEPT_EULA=YES

sqlcmd create mssql get-tags
sqlcmd create mssql --tag 2019-latest
```

To stop, start and delete contexts, run the following commands:

```
sqlcmd stop
sqlcmd start
sqlcmd delete
```

### Backwards compatibility with ODBC sqlcmd

To connect to the current context, and use the original ODBC sqlcmd flags (e.g. -q, -Q, -i, -o etc.), which can be listed with `sqlcmd -?`, run:

```
sqlcmd -q "SELECT @@version"
sqlcmd
```

If no current context exists, `sqlcmd` (with no connection parameters) reverts to the original ODBC `sqlcmd` behavior of creating an interactive session to the default local instance on port 1433 using trusted authentication, otherwise it will create an interactive session to the current context.

## Sqlcmd

The `sqlcmd` project aims to be a complete port of the original ODBC sqlcmd to the `Go` language, utilizing the [go-mssqldb][] driver. For full documentation of the tool and installation instructions, see [go-sqlcmd-utility][].

### Changes in behavior from the ODBC based sqlcmd

- `/` is not accepted as a flag specifier, only `-`
- There are new posix-style versions of each flag, such as `--input-file` for `-i`. `sqlcmd -?` will print those parameter names. Those new names do not preserve backward compatibility with ODBC `sqlcmd`. For example, to specify multiple input file names using `--input-file`, the file names must be comma-delimited, not space-delimited.

The following switches have different behavior in this version of `sqlcmd` compared to the original ODBC based `sqlcmd`.
- `-R` switch is ignored. The go runtime does not provide access to user locale information, and it's not readily available through syscall on all supported platforms.
- `-I` switch is ignored; quoted identifiers are always set on. To disable quoted identifier behavior, add `SET QUOTED IDENTIFIER OFF` in your scripts.
- `-N` now takes an optional string value that can be one of `s[trict]`,`t[rue]`,`m[andatory]`, `yes`,`1`, `o[ptional]`,`no`, `0`, `f[alse]`, or `disable` to specify the encryption choice.
  - If `-N` is passed but no value is provided, `true` is used.
  - If `-N` and `-C` are not provided, sqlcmd will negotiate authentication with the server without validating the server certificate.
  - If `-N` is provided but `-C` is not, sqlcmd will require validation of the server certificate. Note that a `false` value for encryption could still lead to encryption of the login packet.
  - `-C` has no effect when `strict` value is specified for `-N`.
  - If both `-N` and `-C` are provided, sqlcmd will use their values for encryption negotiation.
  - To provide the value of the host name in the server certificate when using strict encryption, pass the host name with `-F`. Example: `-Ns -F myhost.domain.com`
  - More information about client/server encryption negotiation can be found at <https://docs.microsoft.com/openspecs/windows_protocols/ms-tds/60f56408-0188-4cd5-8b90-25c6f2423868>
- `-u` The generated Unicode output file will have the UTF16 Little-Endian Byte-order mark (BOM) written to it.
- Some behaviors that were kept to maintain compatibility with `OSQL` may be changed, such as alignment of column headers for some data types.
- All commands must fit on one line, even `EXIT`. Interactive mode will not check for open parentheses or quotes for commands and prompt for successive lines. The ODBC sqlcmd allows the query run by `EXIT(query)` to span multiple lines.
- `-i` doesn't handle a comma `,` in a file name correctly unless the file name argument is triple quoted. For example:
  `sqlcmd -i """select,100.sql"""` will try to open a file named `sql,100.sql` while `sqlcmd -i "select,100.sql"` will try to open two files `select` and `100.sql`
- If using a single `-i` flag  to pass multiple file names, there must be a space after the `-i`. Example: `-i file1.sql file2.sql`
- `-M` switch is ignored. Sqlcmd always enables multi-subnet failover.

### Switches not available in the new sqlcmd (go-sqlcmd) yet

There are a few switches yet to be implemented in the new `sqlcmd` (go-sqlcmd) compared
to the original ODBC based `sqlcmd`, discussion [#293](https://github.com/microsoft/go-sqlcmd/discussions/292) 
lists these switches. Please provide feedback in the discussion on which 
switches are most important to you to have implemented next in the new sqlcmd.


### Miscellaneous enhancements

- Console output coloring (see below)
- `:Connect` now has an optional `-G` parameter to select one of the authentication methods for Azure SQL Database  - `SqlAuthentication`, `ActiveDirectoryDefault`, `ActiveDirectoryIntegrated`, `ActiveDirectoryServicePrincipal`, `ActiveDirectoryManagedIdentity`, `ActiveDirectoryPassword`. If `-G` is not provided, either Integrated security or SQL Authentication will be used, dependent on the presence of a `-U` username parameter.
- The new `--driver-logging-level` command line parameter allows you to see traces from the `go-mssqldb` client driver. Use `64` to see all traces.
- Sqlcmd can now print results using a vertical format. Use the new `--vertical` command line option to set it. It's also controlled by the `SQLCMDFORMAT` scripting variable.
- Sqlcmd defaults to a horizontal output format (space separated, no borders). To use the new ASCII table format, use the new `--ascii` command line option or set `SQLCMDFORMAT` to `ascii` (`-v SQLCMDFORMAT=ascii`). Note that when using the ASCII table format, individual column widths are determined by the content, but the `SQLCMDCOLWIDTH` variable and the `-w` parameter are still used to control the maximum screen width, determining when columns wrap into separate table segments. The following variables are ignored: `SQLCMDMAXFIXEDTYPEWIDTH`, `SQLCMDMAXVARTYPEWIDTH`, and `SQLCMDHEADERS`.

```
1> select session_id, client_interface_name, program_name from sys.dm_exec_sessions where session_id=@@spid
2> go
session_id            58
client_interface_name go-mssqldb
program_name          sqlcmd
```

- `sqlcmd` supports shared memory and named pipe transport. Use the appropriate protocol prefix on the server name to force a protocol:
  * `lpc` for shared memory, only for a localhost.                           `sqlcmd -S lpc:.`
  * `np` for named pipes. Or use the UNC named pipe path as the server name: `sqlcmd -S \\myserver\pipe\sql\query`
  * `tcp` for tcp                                                            `sqlcmd -S tcp:myserver,1234`
  If no protocol is specified, sqlcmd will attempt to dial in this order: lpc->np->tcp. If dialing a remote host, `lpc` will be skipped.

```
1> select net_transport from sys.dm_exec_connections where session_id=@@spid
2> go
net_transport Named pipe
```

### Azure Active Directory Authentication

`sqlcmd` supports a broader range of AAD authentication models (over the original ODBC based `sqlcmd`), based on the [azidentity package](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity). The implementation relies on an AAD Connector in the [driver](https://github.com/microsoft/go-mssqldb).

#### Command line

To use AAD auth, you can use one of two command line switches:

`-G` is (mostly) compatible with its usage in the prior version of sqlcmd. If a username and password are provided, it will authenticate using AAD Password authentication. If a username is provided it will use AAD Interactive authentication which may display a web browser. If no username or password is provided, it will use a DefaultAzureCredential which attempts to authenticate through a variety of mechanisms.

`--authentication-method=` can be used to specify one of the following authentication types.

`ActiveDirectoryDefault`

- For an overview of the types of authentication this mode will use, see (<https://github.com/Azure/azure-sdk-for-go/tree/main/sdk/azidentity#defaultazurecredential>).
- Choose this method if your database automation scripts are intended to run in both local development environments and in a production deployment in Azure. You'll be able to use a client secret or an Azure CLI login on your development environment and a managed identity or client secret on your production deployment without changing the script.
- Setting environment variables AZURE_TENANT_ID, and AZURE_CLIENT_ID are necessary for DefaultAzureCredential to begin checking the environment configuration and look for one of the following additional environment variables in order to authenticate:

  - Setting environment variable AZURE_CLIENT_SECRET configures the DefaultAzureCredential to choose ClientSecretCredential.
  - Setting environment variable AZURE_CLIENT_CERTIFICATE_PATH configures the DefaultAzureCredential to choose ClientCertificateCredential if AZURE_CLIENT_SECRET is not set.
  - Setting environment variable AZURE_USERNAME configures the DefaultAzureCredential to choose UsernamePasswordCredential if AZURE_CLIENT_SECRET and AZURE_CLIENT_CERTIFICATE_PATH are not set.

`ActiveDirectoryIntegrated`

This method is currently not implemented and will fall back to `ActiveDirectoryDefault`.

`ActiveDirectoryPassword`

This method will authenticate using a username and password. It will not work if MFA is required.
You provide the username and password using the usual command line switches or SQLCMD environment variables.
Set `AZURE_TENANT_ID` environment variable to the tenant id of the server if not using the default tenant of the user.

`ActiveDirectoryInteractive`

This method will launch a web browser to authenticate the user.

`ActiveDirectoryManagedIdentity`

Use this method when running sqlcmd on an Azure VM that has either a system-assigned or user-assigned managed identity. If using a user-assigned managed identity, set the username to the ID of the managed identity. If using a system-assigned identity, leave username empty.

`ActiveDirectoryServicePrincipal`

This method authenticates the provided username as a service principal id and the password as the client secret for the service principal. Provide a username in the form `<service principal id>@<tenant id>`. Set `SQLCMDPASSWORD` variable to the client secret. If using a certificate instead of a client secret, set `AZURE_CLIENT_CERTIFICATE_PATH` environment variable to the path of the certificate file.

#### Environment variables for AAD auth

Some settings for AAD auth do not have command line inputs, and some environment variables are consumed directly by the `azidentity` package used by `sqlcmd`.
These environment variables can be set to configure some aspects of AAD auth and to bypass default behaviors. In addition to the variables listed above, the following are sqlcmd-specific and apply to multiple methods.

`SQLCMDCLIENTID` - set this to the identifier of an application registered in your AAD which is authorized to authenticate to Azure SQL Database. Applies to `ActiveDirectoryInteractive` and `ActiveDirectoryPassword` methods.

### Console colors

Sqlcmd now supports syntax coloring the output of `:list` and the results of TSQL queries when output to the terminal.
To enable coloring use the `SQLCMDCOLORSCHEME` variable, which can be set as an environment variable or by using `:setvar`. The valid values are the names of styles supported by the [chroma styles](https://github.com/alecthomas/chroma/tree/master/styles) project. 

To see a list of available styles along with colored syntax samples, use this command in interactive mode:

```
:list color
```

### Packages

#### sqlcmd executable

Build [sqlcmd](cmd/modern)

```sh
./build/build.sh
```
or
```
.\build\build.cmd
```

#### sqlcmd package

pkg/sqlcmd is consumable by other hosts. Go docs for the package are forthcoming. See the test code and [main.go](cmd/sqlcmd/main.go) for examples of initializing and running sqlcmd.

## Building

```sh
build/build
```

## Testing

The tests rely on SQLCMD scripting variables to provide the connection string parameters. Set SQLCMDSERVER, SQLCMDDATABASE, SQLCMDUSER, SQLCMDPASSWORD variables appropriately then

```sh
go test ./...
```

If you are developing on Windows, you can use docker or WSL to run the tests on Linux. `docker run` lets you pass the environment variables. For example, if your code is in `i:\git\go-sqlcmd` you can run tests in a docker container:

```cmd
docker run -rm -e SQLCMDSERVER=<yourserver> -e SQLCMDUSER=<youruser> -e SQLCMDPASSWORD=<yourpassword> -v i:\git\go-sqlcmd:/go-sqlcmd -w /go-sqlcmd golang:1.16 go test ./...
```

## Localization
The new sqlcmd (go-sqlcmd) is localized for the following languages:
Chinese (Simplified) | Chinese (Traditional) | English (United States) | French | German | Italian | Japanese | Korean | Portuguese (Brazil) | Russian | Spanish

Currently, the user visible strings that also existed in ODBC based sqlcmd are localized in the new sqlcmd, new strings (introduced with the new sqlcmd functionality) will be localized shortly.

To get localized messages from `sqlcmd` set environment variable SQLCMD_LANG to a language tag as per BCP47 convention.
e.g.
```
\git\go-sqlcmd>set SQLCMD_LANG=de-de
\git\go-sqlcmd>.\sqlcmd.exe -w 4
sqlcmd.exe: error: sqlcmd.exe: '-w 4': Der Wert muss größer als 8 und kleiner als 65536 sein.
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
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.

[Community Buzz]: https://github.com/microsoft/go-sqlcmd/discussions/367
[Homebrew]: https://formulae.brew.sh/formula/sqlcmd
[Linuxbrew]: https://docs.brew.sh/Homebrew-on-Linux
[Winget]: https://github.com/microsoft/winget-pkgs/tree/master/manifests/m/Microsoft/Sqlcmd
[Choco]: https://community.chocolatey.org/packages/sqlcmd
[Docker]: https://www.docker.com/products/docker-desktop/
[Podman]: https://podman-desktop.io/downloads/
[releases page]: https://github.com/microsoft/go-sqlcmd/releases/latest
[go-mssqldb]: https://github.com/microsoft/go-mssqldb
[go-sqlcmd-utility]: https://docs.microsoft.com/sql/tools/go-sqlcmd-utility
