# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- `:HELP` command to display available sqlcmd commands
- `-p` flag for printing performance statistics after each batch
- `-j` flag for printing raw error messages without formatting
- `:PERFTRACE` command to redirect timing output to file
- `:SERVERLIST` command to list available SQL Server instances on the network
- Multi-line `EXIT(query)` support in interactive mode - queries with unbalanced parentheses now prompt for continuation

### Fixed
- Statistics format (`-p` flag) now matches ODBC sqlcmd output format
- Panic on empty args slice in command parser

### Changed
- **Breaking for go-sqlcmd users**: `-u` (Unicode output) no longer writes a UTF-16LE BOM (Byte Order Mark) to output files. This change aligns go-sqlcmd with ODBC sqlcmd behavior, which never wrote a BOM. If your workflows depended on the BOM being present, you may need to adjust them.

## Notes on ODBC sqlcmd Compatibility

This release significantly improves compatibility with the original ODBC-based sqlcmd:

| Feature | Previous go-sqlcmd | Now | ODBC sqlcmd |
|---------|-------------------|-----|-------------|
| `-u` output BOM | Wrote BOM | No BOM | No BOM ✓ |
| `-p` statistics format | Different format | Matches | Matches ✓ |
| `-r` without argument | Required argument | Defaults to 0 | Defaults to 0 ✓ |
| `EXIT(query)` multi-line | Not supported | Supported | Supported ✓ |
| `:HELP` command | Not available | Available | Available ✓ |
| `:SERVERLIST` command | Not available | Available | Available ✓ |
