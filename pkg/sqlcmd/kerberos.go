// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/microsoft/go-sqlcmd/internal/localizer"

	// Registers the "krb5" authenticator that -E uses outside Windows
	_ "github.com/microsoft/go-mssqldb/integratedauth/krb5"
)

// kerberosConnectionString adds the krb5 authenticator parameters for -E outside Windows,
// where the driver would otherwise default to NTLM, which needs a user name and password.
func kerberosConnectionString(connectionString string) (string, error) {
	if runtime.GOOS == "windows" {
		return connectionString, nil
	}
	configFile := os.Getenv("KRB5_CONFIG")
	if configFile == "" {
		configFile = "/etc/krb5.conf"
	}
	credCache, err := credentialCachePath(os.Getenv("KRB5CCNAME"))
	if err != nil {
		return "", err
	}
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("authenticator", "krb5")
	query.Set("krb5-configfile", configFile)
	// The driver reads KRBCCNAME, not the MIT standard KRB5CCNAME, so pass the cache explicitly
	query.Set("krb5-credcachefile", credCache)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// credentialCachePath returns the file of a FILE: credential cache name, defaulting to /tmp/krb5cc_<uid>
func credentialCachePath(name string) (string, error) {
	if name == "" {
		return existingCacheFile(fmt.Sprintf("/tmp/krb5cc_%d", os.Getuid()))
	}
	cacheType, path, found := strings.Cut(name, ":")
	if !found || strings.Contains(cacheType, "/") {
		return name, nil
	}
	// The pure Go Kerberos client can't read API: (macOS default), KCM:, KEYRING: or DIR: caches
	if !strings.EqualFold(cacheType, "FILE") {
		return "", localizer.Errorf("Kerberos credential cache type %s is not supported. Use a FILE: cache in KRB5CCNAME.", cacheType)
	}
	return path, nil
}

// existingCacheFile fails early when kinit hasn't written the default cache, instead of the driver's file-not-found error
func existingCacheFile(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", localizer.Errorf("No Kerberos credentials found in %s. Run kinit to get a ticket in a FILE: cache, or set KRB5CCNAME to one.", path)
	}
	return path, nil
}
