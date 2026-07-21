# Changelog

## [1.11.0](https://github.com/microsoft/go-sqlcmd/compare/v1.10.0...v1.11.0) (2026-07-21)


### Features

* add ASCII table query output format ([#609](https://github.com/microsoft/go-sqlcmd/issues/609)) ([91fa9f7](https://github.com/microsoft/go-sqlcmd/commit/91fa9f7b81917913437e6ee3d0c6c6c94ec052bf))
* add devcontainer for VS Code and GitHub Codespaces ([#692](https://github.com/microsoft/go-sqlcmd/issues/692)) ([bdc375d](https://github.com/microsoft/go-sqlcmd/commit/bdc375d32d0eafbdef6392cff5ac1ebc0af24324))
* implement -j raw-errors flag for ODBC sqlcmd compatibility ([#759](https://github.com/microsoft/go-sqlcmd/issues/759)) ([6879335](https://github.com/microsoft/go-sqlcmd/commit/6879335b420c09c27a5749ea0cb353ca42c54b6c))
* remove Azure SQL Edge support ([#680](https://github.com/microsoft/go-sqlcmd/issues/680)) ([24c02ff](https://github.com/microsoft/go-sqlcmd/commit/24c02ff6ed4a234ad9480ea712b944f311655a25))


### Bug Fixes

* reject non-YAML extensions for --sqlconfig flag ([#747](https://github.com/microsoft/go-sqlcmd/issues/747)) ([a185f20](https://github.com/microsoft/go-sqlcmd/commit/a185f203b8d440b536b68a6ba67d50b7e7a67c64))
* use ActiveDirectoryDefault for -G with no username ([#754](https://github.com/microsoft/go-sqlcmd/issues/754)) ([860c66b](https://github.com/microsoft/go-sqlcmd/commit/860c66b5ba63b0ae845b1411a876d869e9a13fa5))
* using ActiveDirectoryServicePrincipalAccessToken does not set password in connection URL ([#756](https://github.com/microsoft/go-sqlcmd/issues/756)) ([d5eb4e7](https://github.com/microsoft/go-sqlcmd/commit/d5eb4e7ed37ebe5711c55e29e4b97a18036e4cb4))
