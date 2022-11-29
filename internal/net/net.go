// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package net

import (
	"net"
	"strconv"
	"time"
)

func IsLocalPortAvailable(port int) (portAvailable bool) {
	timeout := time.Second
	trace(
		"Checking if local port %d is available using DialTimeout(tcp, %v, timeout: %d)",
		port,
		net.JoinHostPort("localhost", strconv.Itoa(port)),
		timeout,
	)
	conn, err := net.DialTimeout(
		"tcp",
		net.JoinHostPort("localhost", strconv.Itoa(port)),
		timeout,
	)
	if err != nil {
		trace(
			"Expected connecting error '%v' on local port %d, therefore port is available)",
			err,
			port,
		)
		portAvailable = true
	}
	if conn != nil {
		err := conn.Close()
		checkErr(err)
		trace("Local port '%d' is not available", port)
	} else {
		trace("Local port '%d' is available", port)
	}

	return
}
