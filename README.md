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

### Packages

#### sqlcmd

#### batch

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

