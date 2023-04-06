package uri

import (
	"path/filepath"
	"strings"
)

func (u Uri) IsLocal() bool {
	if len(u.Scheme()) > 2 {
		return false
	} else {
		return true
	}
}

// actualUrl returns the url without the query string
func (u Uri) ActualUrl() string {
	terminator := strings.LastIndex(u.uri, ",")
	if terminator != -1 {
		return u.uri[0:terminator]
	} else {
		return u.uri
	}
}

func (u Uri) Scheme() string {
	return u.url.Scheme
}

func (u Uri) FileExtension() string {
	_, f := filepath.Split(u.ActualUrl())
	return strings.TrimLeft(filepath.Ext(f), ".")
}

func (u Uri) Filename() string {
	filename := filepath.Base(u.ActualUrl())
	if filename == "" {
		panic("filename is empty")
	}
	return filename
}

// parseDbName returns the databaseName from --using arg
// It sets database name to the specified database name
// or in absence of it, it is set to the filename without
// extension.
func (u Uri) ParseDbName() string {
	if u.uri == "" {
		panic("uri is empty")
	}

	// TODO: Reimplement
	return ""
}

func (u Uri) GetDbNameAsIdentifier() string {
	escapedDbName := strings.ReplaceAll(u.ParseDbName(), "'", "''")
	dbName := strings.ReplaceAll(escapedDbName, "]", "]]")
	if dbName == "" {
		panic("database name is empty")
	}
	return dbName
}

func (u Uri) GetDbNameAsNonIdentifier() string {
	dbName := strings.ReplaceAll(u.ParseDbName(), "]", "]]")
	if dbName == "" {
		panic("database name is empty")
	}
	return dbName
}
