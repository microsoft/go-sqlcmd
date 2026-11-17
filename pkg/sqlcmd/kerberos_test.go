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
	if runtime.GOOS == "windows" {
		t.Skip("-E uses SSPI on Windows")
	}
	dir := t.TempDir()
	cache := filepath.Join(dir, "cc")
	colonCache := filepath.Join(dir, "a:b")
	for _, path := range []string{cache, colonCache} {
		require.NoError(t, os.WriteFile(path, nil, 0o600))
	}
	tests := []struct {
		name, ccname, path, errText string
	}{
		{"FILE prefix", "FILE:" + cache, cache, ""},
		{"lowercase prefix", "file:" + cache, cache, ""},
		{"no prefix", cache, cache, ""},
		{"colon in path", colonCache, colonCache, ""},
		{"missing file", "FILE:" + filepath.Join(dir, "missing"), "", "No Kerberos credentials found in"},
		{"empty FILE path", "FILE:", "", "No Kerberos credentials found in"},
		{"API", "API:1234", "", "is not supported"},
		{"KCM", "KCM:1000", "", "is not supported"},
		{"KEYRING", "KEYRING:persistent:1000", "", "is not supported"},
		{"DIR", "DIR:" + dir, "", "is not supported"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, err := credentialCachePath(test.ccname)
			if test.errText != "" {
				assert.ErrorContains(t, err, test.errText)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.path, path)
		})
	}
}

func TestKerberosAuthenticator(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("-E uses SSPI on Windows")
	}
	dir := t.TempDir()
	configFile := filepath.Join(dir, "krb5.conf")
	cache := filepath.Join(dir, "krb5cc_test")
	require.NoError(t, os.WriteFile(configFile, []byte("[libdefaults]\n default_realm = EXAMPLE.COM\n"), 0o600))
	require.NoError(t, os.WriteFile(cache, nil, 0o600))
	defaultCache := fmt.Sprintf("/tmp/krb5cc_%d", os.Getuid())
	_, statErr := os.Stat(defaultCache)
	defaultExists := statErr == nil
	defaultErr := "No Kerberos credentials found in " + defaultCache
	if defaultExists {
		defaultErr = ""
	}
	tests := []struct {
		name, ccname, wantCache, errText string
	}{
		{"FILE cache", "FILE:" + cache, cache, ""},
		{"plain path", cache, cache, ""},
		{"missing cache", "FILE:" + filepath.Join(dir, "missing"), "", "No Kerberos credentials found in"},
		{"unsupported type", "API:x", "", "is not supported"},
		{"default cache", "", defaultCache, defaultErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("KRB5_CONFIG", configFile)
			t.Setenv("KRB5CCNAME", test.ccname)
			base, err := ConnectSettings{UseTrustedConnection: true, ServerName: "sqlhost"}.ConnectionString()
			require.NoError(t, err)
			connstr, err := kerberosConnectionString(base)
			if test.errText != "" {
				require.ErrorContains(t, err, test.errText)
				return
			}
			require.NoError(t, err)
			config, err := msdsn.Parse(connstr)
			require.NoError(t, err)
			assert.Equal(t, "krb5", config.Parameters["authenticator"])
			assert.Equal(t, test.wantCache, config.Parameters["krb5-credcachefile"])
			auth, err := integratedauth.GetIntegratedAuthenticator(config)
			require.NoError(t, err)
			assert.NotNil(t, auth)
		})
	}
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
		{"no -E and no user", ConnectSettings{}, false, true, true},
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
