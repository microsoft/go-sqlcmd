package ingest

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/databaseurl"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/location"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"
)

type ingest struct {
	url         *databaseurl.DatabaseUrl
	location    location.Location
	controller  *container.Controller
	mechanism   mechanism.Mechanism
	options     mechanism.BringOnlineOptions
	containerId string
	query       func(text string)
}

func (i *ingest) IsRemoteUrl() bool {
	return !i.location.IsLocal()
}

func (i *ingest) UrlFilename() string {
	return i.url.Filename
}

func (i *ingest) OnlineMethod() string {
	return i.mechanism.Name()
}

func (i *ingest) DatabaseName() string {
	return i.url.DatabaseName
}

func (i *ingest) IsValidScheme() bool {
	for _, s := range i.location.ValidSchemes() {
		if s == i.url.Scheme {
			return true
		}
	}
	return false
}

func (i *ingest) CopyToContainer(containerId string) {
	destFolder := "/var/opt/mssql/backup"

	if i.mechanism != nil {
		destFolder = i.mechanism.CopyToLocation()
	}
	if i.location == nil {
		panic("location is nil, did you call NewIngest()?")
	}

	i.containerId = containerId
	i.location.CopyToContainer(containerId, destFolder)
	i.options.Filename = i.url.Filename

	if i.options.Filename == "" {
		panic("filename is empty")
	}
}

func (i *ingest) BringOnline(query func(string), username string, password string) {
	if i.options.Filename == "" {
		panic("filename is empty, did you call CopyToContainer()?")
	}
	if query == nil {
		panic("query is nil")
	}
	if i.mechanism == nil {
		panic("mechanism is nil")
	}

	i.query = query
	i.options.Username = username
	i.options.Password = password
	i.mechanism.BringOnline(i.url.DatabaseNameAsTsqlIdentifier, i.containerId, i.query, i.options)
	i.setDefaultDatabase(username)
}

func (i *ingest) setDefaultDatabase(username string) {
	if i.query == nil {
		panic("query is nil, did you call BringOnline()?")
	}

	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		username,
		i.url.DatabaseNameAsNonTsqlIdentifier)
	i.query(alterDefaultDb)
}

func (i *ingest) IsValidFileExtension() bool {
	for _, m := range mechanism.FileTypes() {
		if m == i.url.FileExtension {
			return true
		}
	}
	return false
}

func (i *ingest) SourceFileExists() bool {
	return i.location.Exists()
}

func (i *ingest) UserProvidedFileExt() string {
	return i.url.FileExtension
}

func (i *ingest) ValidSchemes() []string {
	return i.location.ValidSchemes()
}
