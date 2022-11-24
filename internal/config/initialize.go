// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/net"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/secret"
)

var encryptCallback func(plainText string, encrypt bool) (cipherText string)
var decryptCallback func(cipherText string, decrypt bool) (secret string)
var isLocalPortAvailableCallback func(port int) (portAvailable bool)

func init() {
	errorHandler := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	Initialize(
		errorHandler,
		output.Tracef,
		secret.Encode,
		secret.Decode,
		net.IsLocalPortAvailable)
}

func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	encryptHandler func(plainText string, encrypt bool) (cipherText string),
	decryptHandler func(cipherText string, decrypt bool) (secret string),
	isLocalPortAvailableHandler func(port int) (portAvailable bool),
) {
	errorCallback = errorHandler
	traceCallback = traceHandler
	encryptCallback = encryptHandler
	decryptCallback = decryptHandler
	isLocalPortAvailableCallback = isLocalPortAvailableHandler
}
