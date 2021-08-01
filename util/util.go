// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package util

import (
	"strconv"
	"strings"

	"github.com/microsoft/go-sqlcmd/sqlcmderrors"
)

func SplitServer(serverName string) (server string, instance string, port uint64, err error) {
	if strings.HasPrefix(serverName, "tcp:") {
		if len(serverName) == 4 {
			return "", "", 0, &sqlcmderrors.InvalidServerName
		}
		serverName = serverName[4:]
	}
	serverNameParts := strings.Split(serverName, ",")
	if len(serverNameParts) > 2 {
		return "", "", 0, &sqlcmderrors.InvalidServerName
	}
	if len(serverNameParts) == 2 {
		var err error
		port, err = strconv.ParseUint(serverNameParts[1], 10, 16)
		if err != nil {
			return "", "", 0, &sqlcmderrors.InvalidServerName
		}
		serverName = serverNameParts[0]
	} else {
		serverNameParts = strings.Split(serverName, "/")
		if len(serverNameParts) > 2 {
			return "", "", 0, &sqlcmderrors.InvalidServerName
		}
		if len(serverNameParts) == 2 {
			instance = serverNameParts[1]
			serverName = serverNameParts[0]
		}
	}
	return serverName, instance, port, nil
}

func PadRight(builder *strings.Builder, c int64, s string) *strings.Builder {
	var i int64
	for ; i < c; i++ {
		builder.WriteString(s)
	}
	return builder
}

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
