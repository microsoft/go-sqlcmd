# SQL Utilities - Go edition

This repo contains command line tools and go packages for working with Microsoft SQL Server, Azure SQL Database, and Azure Synapse.

## Sqlcmd

The `sqlcmd` project aims to be a complete port of the native sqlcmd to the `go` language, utilizing the [go-mssqldb](https://github.com/denisenkom/go-mssqldb) driver. For full documentation of the tool, see https://docs.microsoft.com/sql/tools/sqlcmd-utility

### Breaking changes

We will be implementing as many command line switches and behaviors as possible over time. Several switches and behaviors are expected to change in this implementation.

- `-P` switch will be removed. Passwords for SQL authentication can only be provided through these mechanisms:

    -The `SQLCMDPASSWORD` environment variable
    -The `:CONNECT` command
    -When prompted, the user can type the password to complete a connection

- `-R` switch will be removed. The go runtime does not provide access to user locale information, and it's not readily available through syscall on all supported platforms.
- Some behaviors that were kept to maintain compatibility with `OSQL` may be changed, such as alignment of column headers for some data types.
- All commands must fit on one line, even `EXIT`. Interactive mode will not check for open parentheses or quotes for commands and prompt for successive lines. The native sqlcmd allows the query run by `EXIT(query)` to span multiple lines.

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

