#!/bin/bash

echo "Starting SQL Server in background"
/opt/mssql/bin/sqlservr &

# wait for SQL Server to start up.
sleep 30s

echo "running script with SQLCMD legacy"

/opt/mssql-tools/bin/sqlcmd -i /app/myscript.sql -U SA

echo "running script with SQLCMD GO"

sqlcmd-go -i /app/myscript.sql -U SA