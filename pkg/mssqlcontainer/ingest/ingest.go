package ingest

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/databaseurl"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/extract"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/location"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"
	"path/filepath"
	"strings"
)

type ingest struct {
	url         *databaseurl.DatabaseUrl
	location    location.Location
	controller  *container.Controller
	mechanism   mechanism.Mechanism
	options     mechanism.BringOnlineOptions
	extractor   extract.Extractor
	containerId string
	query       func(text string)
}

func (i *ingest) IsExtractionNeeded() bool {
	i.extractor = extract.NewExtractor(i.url.FileExtension, i.controller)
	if i.extractor == nil {
		return false
	} else {
		return true
	}
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

func (i *ingest) Extract() {
	if i.extractor == nil {
		panic("extractor is nil")
	}

	if !i.extractor.IsInstalled(i.containerId) {
		i.extractor.Install()
	}

	i.options.Filename, i.options.LdfFilename =
		i.extractor.Extract(i.url.Filename, "/var/opt/mssql/data")

	if i.mechanism == nil {
		ext := strings.TrimLeft(filepath.Ext(i.options.Filename), ".")
		i.mechanism = mechanism.NewMechanismByFileExt(ext, i.controller)
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

	if i.mechanism.Name() != "git" {
		i.setDefaultDatabase(username)
	}
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
	for _, e := range extract.FileTypes() {
		if e == i.url.FileExtension {
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
