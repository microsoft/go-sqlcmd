package databaseurl

import (
	"fmt"
	url2 "net/url"
	"path/filepath"
	"strings"
)

func NewDatabaseUrl(url string) *DatabaseUrl {
	databaseUrl := DatabaseUrl{}

	// To enable URL.Parse, switch to / from \\
	url = strings.Replace(url, "\\", "/", -1)

	// Cope with a URL that has no schema, e.g. in local dir, or local folder \foo.mdf
	if !strings.Contains(url, "/") {
		url = "./" + url
	}

	var err error
	parsedUrl, err := url2.Parse(url)
	if err != nil {
		panic(err)
	}

	databaseUrl.URL = parsedUrl

	fmt.Println("databaseUrl.URL.Path: " + databaseUrl.URL.Path)

	databaseUrl.Filename = filepath.Base(databaseUrl.URL.Path)
	databaseUrl.FileExtension = strings.TrimLeft(filepath.Ext(databaseUrl.Filename), ".")

	split := strings.Split(databaseUrl.URL.Path, ",")
	if len(split) > 1 {
		databaseUrl.DatabaseName = split[1]

		// Remove the database name (specified after the comma)  from the URL, and reparse it
		url = strings.Replace(url, ","+split[1], "", 1)
		databaseUrl.URL, err = databaseUrl.URL.Parse(url)
		if err != nil {
			panic(err)
		}

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

	fmt.Println("databaseUrl.Filename: " + databaseUrl.Filename)
	fmt.Println("databaseUrl.FileExtension: " + databaseUrl.FileExtension)
	fmt.Println("databaseUrl.DatabaseName: " + databaseUrl.DatabaseName)

	databaseUrl.IsLocal = databaseUrl.URL.Scheme == "file" || len(databaseUrl.URL.Scheme) < 3

	escapedDbName := strings.ReplaceAll(databaseUrl.DatabaseName, "'", "''")
	databaseUrl.DatabaseNameAsTsqlIdentifier = strings.ReplaceAll(escapedDbName, "]", "]]")
	databaseUrl.DatabaseNameAsNonTsqlIdentifier = strings.ReplaceAll(databaseUrl.DatabaseName, "]", "]]")

	return &databaseUrl
}
