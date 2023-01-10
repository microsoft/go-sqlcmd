#!/bin/sh
scriptdir=`dirname $0`
go build -o $scriptdir/../sqlcmd $scriptdir/../cmd/modern
