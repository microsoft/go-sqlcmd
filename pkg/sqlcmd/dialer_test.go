// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyDialerHostName(t *testing.T) {
	d := &proxyDialer{serverName: "myserver.database.windows.net"}
	assert.Equal(t, "myserver.database.windows.net", d.HostName())
}

func TestProxyDialerHostNameEmpty(t *testing.T) {
	d := &proxyDialer{}
	assert.Equal(t, "", d.HostName())
}

func TestProxyDialerInitializesNetDialer(t *testing.T) {
	d := &proxyDialer{serverName: "test.server.net"}
	assert.Nil(t, d.dialer)

	// DialContext should fail with an invalid address, but that's fine for this test
	// We just want to verify the dialer gets initialized
	_, _ = d.DialContext(context.Background(), "tcp", "invalid:99999")
	assert.NotNil(t, d.dialer)
}

func TestProxyDialerDialAddressOverridesHostAndPortForTCP(t *testing.T) {
	d := &proxyDialer{
		targetHost: "proxy.local",
		targetPort: "1444",
	}

	dialAddr := d.dialAddress("tcp", "server.example.com:1433")
	assert.Equal(t, "proxy.local:1444", dialAddr)
}

func TestProxyDialerDialAddressKeepsPortForUDP(t *testing.T) {
	d := &proxyDialer{
		targetHost: "proxy.local",
		targetPort: "1444",
	}

	dialAddr := d.dialAddress("udp", "server.example.com:1434")
	assert.Equal(t, "proxy.local:1434", dialAddr)
}
