package databaseurl

import (
	url2 "net/url"
	"path/filepath"
	"strings"
)

func NewDatabaseUrl(url string) *DatabaseUrl {
	trace("NewDatabaseUrl(" + url + ")")

	databaseUrl := DatabaseUrl{}

	// To enable URL.Parse, switch to / from \\
	url = strings.Replace(url, "\\", "/", -1)

	// Cope with a URL that in the local directory, so it can be URL.Parsed()
	if !strings.Contains(url, "/") {
		url = "./" + url
	}

	parsedUrl, err := url2.Parse(url)
	checkErr(err)

	databaseUrl.URL = parsedUrl

	trace("databaseUrl.URL.Path: " + databaseUrl.URL.Path)

	databaseUrl.Filename = filepath.Base(databaseUrl.URL.Path)
	databaseUrl.FileExtension = strings.TrimLeft(filepath.Ext(databaseUrl.Filename), ".")

	split := strings.Split(databaseUrl.URL.Path, ",")
	if len(split) > 1 {
		databaseUrl.DatabaseName = split[1]

		// Remove the database name (specified after the comma)  from the URL, and reparse it
		url = strings.Replace(url, ","+split[1], "", 1)
		databaseUrl.URL, err = databaseUrl.URL.Parse(url)
		checkErr(err)

		split := strings.Split(databaseUrl.FileExtension, ",")
		databaseUrl.FileExtension = split[0]

		split = strings.Split(databaseUrl.Filename, ",")
		databaseUrl.Filename = split[0]
	} else {

		databaseUrl.DatabaseName = strings.TrimSuffix(
			databaseUrl.Filename,
			"."+databaseUrl.FileExtension,
		)
	}

	trace("databaseUrl.Filename: " + databaseUrl.Filename)
	trace("databaseUrl.FileExtension: " + databaseUrl.FileExtension)
	trace("databaseUrl.DatabaseName: " + databaseUrl.DatabaseName)

	databaseUrl.IsLocal = databaseUrl.URL.Scheme == "file" || len(databaseUrl.URL.Scheme) < 3

	escapedDbName := strings.ReplaceAll(databaseUrl.DatabaseName, "'", "''")
	databaseUrl.DatabaseNameAsTsqlIdentifier = strings.ReplaceAll(escapedDbName, "]", "]]")
	databaseUrl.DatabaseNameAsNonTsqlIdentifier = strings.ReplaceAll(databaseUrl.DatabaseName, "]", "]]")

	return &databaseUrl
}
