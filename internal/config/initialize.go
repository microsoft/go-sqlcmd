// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/net"
	"github.com/microsoft/go-sqlcmd/internal/secret"
)

var encryptCallback func(plainText string, encryptionMethod string) (cipherText string)
var decryptCallback func(cipherText string, encryptionMethod string) (secret string)
var isLocalPortAvailableCallback func(port int) (portAvailable bool)

// init sets up the package to work with a set of handlers to be used for the period
// before the command-line has been parsed
func init() {
	errorHandler := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	traceHandler := func(format string, a ...any) {
		fmt.Printf(format, a...)
	}

	Initialize(
		errorHandler,
		traceHandler,
		secret.Encode,
		secret.Decode,
		net.IsLocalPortAvailable)
}

// Initialize sets the callback functions used by the config package.
// These callback functions are used for logging errors, tracing debug messages,
// encrypting and decrypting data, and checking if a local port is available.
// The callback functions are passed to the function as arguments.
// This function should be called at the start of the application to ensure that the
// config package has the necessary callback functions available.
func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	encryptHandler func(plainText string, encryptionMethod string) (cipherText string),
	decryptHandler func(cipherText string, encryptionMethod string) (secret string),
	isLocalPortAvailableHandler func(port int) (portAvailable bool),
) {
	errorCallback = errorHandler
	traceCallback = traceHandler
	encryptCallback = encryptHandler
	decryptCallback = decryptHandler
	isLocalPortAvailableCallback = isLocalPortAvailableHandler
}
