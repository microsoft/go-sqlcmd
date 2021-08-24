// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"strconv"
	"strings"
)

// SplitServer extracts connection parameters from a server name input
func SplitServer(serverName string) (string, string, uint64, error) {
	instance := ""
	port := uint64(0)
	if strings.HasPrefix(serverName, "tcp:") {
		if len(serverName) == 4 {
			return "", "", 0, &InvalidServerName
		}
		serverName = serverName[4:]
	}
	serverNameParts := strings.Split(serverName, ",")
	if len(serverNameParts) > 2 {
		return "", "", 0, &InvalidServerName
	}
	if len(serverNameParts) == 2 {
		var err error
		port, err = strconv.ParseUint(serverNameParts[1], 10, 16)
		if err != nil {
			return "", "", 0, &InvalidServerName
		}
		serverName = serverNameParts[0]
	} else {
		serverNameParts = strings.Split(serverName, "/")
		if len(serverNameParts) > 2 {
			return "", "", 0, &InvalidServerName
		}
		if len(serverNameParts) == 2 {
			instance = serverNameParts[1]
			serverName = serverNameParts[0]
		}
	}
	return serverName, instance, port, nil
}

// PadRight appends c instances of s to builder
func PadRight(builder *strings.Builder, c int64, s string) *strings.Builder {
	var i int64
	for ; i < c; i++ {
		builder.WriteString(s)
	}
	return builder
}

// PadLeft prepends c instances of s to builder
func PadLeft(builder *strings.Builder, c int64, s string) *strings.Builder {
	newBuilder := new(strings.Builder)
	newBuilder.Grow(builder.Len())
	var i int64
	for ; i < c; i++ {
		newBuilder.WriteString(s)
	}
	newBuilder.WriteString(builder.String())
	return newBuilder
}
