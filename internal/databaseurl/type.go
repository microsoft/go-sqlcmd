package databaseurl

import "net/url"

type DatabaseUrl struct {
	*url.URL

	Filename string

	// Is this .git or git!
	FileExtension string
	IsLocal       bool

	// DatabaseName returns the databaseName from --use arg
	// It sets database name to the specified database name
	// or in absence of it, it is set to the filename without
	// extension.
	DatabaseName string

	DatabaseNameAsTsqlIdentifier    string
	DatabaseNameAsNonTsqlIdentifier string
}
