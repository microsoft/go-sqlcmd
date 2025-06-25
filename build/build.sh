#!/bin/sh
scriptdir=`dirname $0`
versionTag=`git describe --tags --abbrev=0`
go build -o $scriptdir/../sqlcmd -ldflags="-X main.version=$versionTag" $scriptdir/../cmd/modern

go install github.com/google/go-licenses@latest
go-licenses report github.com/microsoft/go-sqlcmd/cmd/modern --template build/NOTICE.tpl --ignore github.com/microsoft > $scriptdir/notice.txt
cat $scriptdir/NOTICE.header $scriptdir/notice.txt > $scriptdir/../NOTICE.md
rm $scriptdir/notice.txt
