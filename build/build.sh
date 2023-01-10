#!/bin/sh
scriptdir=`dirname $0`
versionTag=`git describe --tags --abbrev=0`
go build -o $scriptdir/../sqlcmd -ldflags="-X main.version=$versionTag" $scriptdir/../cmd/modern
