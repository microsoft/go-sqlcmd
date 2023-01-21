#!/bin/bash

echo "Starting SQL Server in background"
/opt/mssql/bin/sqlservr &

# wait for SQL Server to start up.
sleep 30s

sqlcmd -i /app/myscript.sql -U SA