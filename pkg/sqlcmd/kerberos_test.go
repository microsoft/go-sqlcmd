// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/microsoft/go-mssqldb/integratedauth"
	"github.com/microsoft/go-mssqldb/msdsn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialCachePath(t *testing.T) {
	tests := []struct {
		name, ccname, path string
		wantErr            bool
	}{
		{"FILE prefix", "FILE:/tmp/cc", "/tmp/cc", false},
		{"lowercase prefix", "file:/tmp/cc", "/tmp/cc", false},
		{"no prefix", "/tmp/cc", "/tmp/cc", false},
		{"colon in path", "/tmp/a:b", "/tmp/a:b", false},
		{"API", "API:1234", "", true},
		{"KCM", "KCM:1000", "", true},
		{"KEYRING", "KEYRING:persistent:1000", "", true},
		{"DIR", "DIR:/tmp/cc", "", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, err := credentialCachePath(test.ccname)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.path, path)
		})
	}
}

func TestExistingCacheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "krb5cc_test")
	_, err := existingCacheFile(path)
	require.ErrorContains(t, err, "No Kerberos credentials found in "+path)

	require.NoError(t, os.WriteFile(path, nil, 0o600))
	got, err := existingCacheFile(path)
	assert.NoError(t, err)
	assert.Equal(t, path, got)
}

func TestKerberosConnectionString(t *testing.T) {
	t.Setenv("KRB5_CONFIG", "/tmp/krb5.conf")
	t.Setenv("KRB5CCNAME", "FILE:/tmp/cc")
	connect := ConnectSettings{UseTrustedConnection: true, ServerName: "sqlhost,1433"}
	base, err := connect.ConnectionString()
	require.NoError(t, err)
	connstr, err := kerberosConnectionString(base)
	require.NoError(t, err)
	if runtime.GOOS == "windows" {
		assert.Equal(t, base, connstr)
		return
	}
	config, err := msdsn.Parse(connstr)
	require.NoError(t, err)
	assert.Equal(t, "krb5", config.Parameters["authenticator"])
	assert.Equal(t, "/tmp/krb5.conf", config.Parameters["krb5-configfile"])
	assert.Equal(t, "/tmp/cc", config.Parameters["krb5-credcachefile"])

	t.Setenv("KRB5_CONFIG", "")
	connstr, err = kerberosConnectionString(base)
	require.NoError(t, err)
	config, err = msdsn.Parse(connstr)
	require.NoError(t, err)
	assert.Equal(t, "/etc/krb5.conf", config.Parameters["krb5-configfile"])

	t.Setenv("KRB5CCNAME", "API:1234")
	_, err = kerberosConnectionString(base)
	assert.Error(t, err)
}

func TestKrb5ProviderRegistered(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("-E uses SSPI on Windows")
	}
	config, err := msdsn.Parse("sqlserver://sqlhost?authenticator=krb5&krb5-configfile=/tmp/krb5.conf&krb5-credcachefile=/tmp/cc")
	require.NoError(t, err)
	_, err = integratedauth.GetIntegratedAuthenticator(config)
	assert.NotContains(t, fmt.Sprint(err), "unknown authenticator")
}

func TestConnectDbKerberos(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("-E uses SSPI on Windows")
	}
	t.Setenv("KRB5CCNAME", "API:x")
	const unsupported = "is not supported"
	tests := []struct {
		name       string
		settings   ConnectSettings
		useCurrent bool
		kerberos   bool
		printed    bool
	}{
		{"-E alone", ConnectSettings{UseTrustedConnection: true}, false, true, true},
		{"-E on current connection", ConnectSettings{UseTrustedConnection: true}, true, true, false},
		{"-E with SqlPassword", ConnectSettings{UseTrustedConnection: true, AuthenticationMethod: SqlPassword, UserName: "u", Password: "p"}, false, false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.settings.ServerName = "127.0.0.1:1"
			s := New(nil, "", InitializeVariables(false))
			buf := &memoryBuffer{buf: new(bytes.Buffer)}
			s.SetOutput(buf)
			var err error
			if test.useCurrent {
				s.Connect = &test.settings
				err = s.ConnectDb(nil, true)
			} else {
				err = s.ConnectDb(&test.settings, true)
			}
			require.Error(t, err)
			assert.Equal(t, test.kerberos, strings.Contains(err.Error(), unsupported), "error: %v", err)
			assert.Equal(t, test.printed, strings.Contains(buf.buf.String(), unsupported), "output: %q", buf.buf.String())
		})
	}
}
