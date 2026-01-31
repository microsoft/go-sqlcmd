// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"net"
	"strings"
)

// proxyDialer implements mssql.HostDialer to allow specifying a server name
// for the TDS login packet that differs from the dial address. This enables
// tunneling connections through localhost while authenticating to the real server.
type proxyDialer struct {
	serverName string
	targetHost string
	targetPort string
	dialer     *net.Dialer
}

func (d *proxyDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if d.dialer == nil {
		d.dialer = &net.Dialer{}
	}
	return d.dialer.DialContext(ctx, network, d.dialAddress(network, addr))
}

func (d *proxyDialer) HostName() string {
	return d.serverName
}

func (d *proxyDialer) dialAddress(network, addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}

	if d.targetHost != "" {
		host = d.targetHost
	}
	if d.targetPort != "" && isTCPNetwork(network) {
		port = d.targetPort
	}

	return net.JoinHostPort(host, port)
}

func isTCPNetwork(network string) bool {
	return strings.HasPrefix(network, "tcp")
}
