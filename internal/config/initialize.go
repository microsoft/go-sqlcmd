// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/microsoft/go-sqlcmd/internal/file"
	"github.com/microsoft/go-sqlcmd/internal/net"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"os"
	"path/filepath"
)

var encryptCallback func(plainText string, encrypt bool) (cipherText string)
var decryptCallback func(cipherText string, decrypt bool) (secret string)
var isLocalPortAvailableCallback func(port int) (portAvailable bool)
var createEmptyFileIfNotExistsCallback func(filename string)

func init() {
	home, _ := os.UserHomeDir()
	configFile := filepath.Join(home, ".sqlcmd", "sqlconfig")

	Initialize(
		func(err error) {
			if err != nil {
				panic(err)
			}
		},
		output.Tracef,
		secret.Encode,
		secret.Decode,
		net.IsLocalPortAvailable,
		file.CreateEmptyIfNotExists,
		configFile,
	)
}

func Initialize(
	errorHandler func(err error),
	traceHandler func(format string, a ...any),
	encryptHandler func(plainText string, encrypt bool) (cipherText string),
	decryptHandler func(cipherText string, decrypt bool) (secret string),
	isLocalPortAvailableHandler func(port int) (portAvailable bool),
	createEmptyFileIfNotExistsHandler func(filename string),
	configFile string,
) {
	errorCallback = errorHandler
	traceCallback = traceHandler
	encryptCallback = encryptHandler
	decryptCallback = decryptHandler
	isLocalPortAvailableCallback = isLocalPortAvailableHandler
	createEmptyFileIfNotExistsCallback = createEmptyFileIfNotExistsHandler

	configureViper(configFile)
	load()
}
