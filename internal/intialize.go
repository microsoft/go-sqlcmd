// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package internal

import (
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/file"
	"github.com/microsoft/go-sqlcmd/internal/mssql"
	"github.com/microsoft/go-sqlcmd/internal/net"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"os"
)

type initInfo struct {
	ErrorHandler func(error)
	TraceHandler func(string, ...any)
}

func Initialize(
	errorHandler func(error),
	hintHandler func([]string),
	sqlconfigFilename string,
	outputType string,
	loggingLevel int,
) {
	info := initInfo{errorHandler, output.Tracef}

	file.Initialize(info.ErrorHandler, info.TraceHandler)
	mssql.Initialize(info.ErrorHandler, info.TraceHandler, secret.Decode)
	output.Initialize(info.ErrorHandler, info.TraceHandler, hintHandler, os.Stdout, outputType, verbosity.Enum(loggingLevel))
	config.Initialize(info.ErrorHandler, info.TraceHandler, secret.Encode, secret.Decode, net.IsLocalPortAvailable, file.CreateEmptyIfNotExists, sqlconfigFilename)
	container.Initialize(info.ErrorHandler, info.TraceHandler)
	secret.Initialize(info.ErrorHandler)
	net.Initialize(info.ErrorHandler, info.TraceHandler)
	pal.Initialize(info.ErrorHandler)
}
