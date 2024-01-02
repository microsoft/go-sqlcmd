// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package net

import (
	"net"
	"strconv"
	"time"
)

// IsLocalPortAvailable takes a port number and returns a boolean indicating
// whether the port is available for use.
func IsLocalPortAvailable(port int) (portAvailable bool) {
	timeout := time.Second

	hostPort := net.JoinHostPort("localhost", strconv.Itoa(port))
	trace(
		"Checking if local port %#v is available using DialTimeout(tcp, %v, timeout: %d)",
		port,
		hostPort,
		timeout,
	)
	conn, err := net.DialTimeout(
		"tcp",
		hostPort,
		timeout,
	)
	if err != nil {
		trace(
			"Expected connecting error '%v' on local port %#v, therefore port is available)",
			err,
			port,
		)
		portAvailable = true
	}
	if conn != nil {
		err := conn.Close()
		checkErr(err)
		trace("Local port '%#v' is not available", port)
	} else {
		trace("Local port '%#v' is available", port)
	}

	return
}
