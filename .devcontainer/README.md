# go-sqlcmd Development Container

Dev Container / Codespaces environment with Go 1.24 and SQL Server 2025.

## Quick Start

**VS Code**: Open repo → click "Reopen in Container" when prompted.

**Codespaces**: Click "Code" → "Codespaces" → "Create codespace".

First build takes ~5 minutes.

## Commands

| Alias | What it does |
|-------|--------------|
| `gtest` | Run tests |
| `ginstall` | Build and install sqlcmd to ~/bin |
| `glint` | Run golangci-lint |
| `sql` | Connect to SQL Server (go-sqlcmd) |
| `sql-legacy` | Connect with legacy ODBC sqlcmd |
| `test-db` | Test database connection |

## SQL Server Connection

- **Server**: `localhost,1433`
- **User**: `sa`
- **Password**: `$SQLCMDPASSWORD` env var (`SqlCmd@2025!` for local dev)
- **Database**: `master` or `SqlCmdTest`

Port 1433 is forwarded — connect from host tools (ADS, SSMS) using same credentials.

## Two sqlcmd Versions

- **go-sqlcmd**: `~/bin/sqlcmd` (default in PATH, use `sql` alias)
- **Legacy ODBC**: `/opt/mssql-tools18/bin/sqlcmd` (use `sql-legacy` alias)

## Customization

**Change SQL version**: Edit `docker-compose.yml` image tag.

**Add setup scripts**: Edit `.devcontainer/mssql/setup.sql`.

**Change password**: Update `docker-compose.yml` and `devcontainer.json`.

## Troubleshooting

- **ARM64 (Apple Silicon)**: Use GitHub Codespaces instead - SQL Server has no native ARM64 images
- **SQL Server not starting**: Check `docker logs $(docker ps -qf "name=db")`. Needs 2GB+ RAM
- **Connection refused**: Wait ~30s for SQL Server to start
- **sqlcmd not found**: Run `ginstall`
