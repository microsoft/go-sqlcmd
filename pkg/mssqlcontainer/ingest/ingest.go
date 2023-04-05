package ingest

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/uri"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/extract"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/location"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"
	"path/filepath"
)

type ingest struct {
	uri         uri.Uri
	location    location.Location
	mechanism   mechanism.Mechanism
	options     mechanism.BringOnlineOptions
	extractor   extract.Extractor
	containerId string
	query       func(text string)
}

func (i ingest) IsExtractionNeeded() bool {
	i.extractor = extract.NewExtractor(i.uri.FileExtension())
	if i.extractor == nil {
		return false
	} else {
		return true
	}
}

func (i ingest) IsRemoteUrl() bool {
	return !i.location.IsLocal()
}

func (i ingest) IsValidScheme() bool {
	for _, s := range i.location.ValidSchemes() {
		if s == i.uri.Scheme() {
			return true
		}
	}
	return false
}

func (i ingest) CopyToLocation() string {
	return i.mechanism.CopyToLocation()
}

func (i ingest) CopyToContainer(containerId string) {
	i.containerId = containerId
	i.location.CopyToContainer(containerId, i.CopyToLocation())
	i.options.Filename = i.uri.Filename()
}

func (i ingest) Extract() {
	if i.extractor == nil {
		panic("extractor is nil")
	}

	if !i.extractor.IsInstalled(i.containerId) {
		i.extractor.Install()
	}

	i.options.Filename, i.options.LdfFilename =
		i.extractor.Extract(i.uri.ActualUrl(), i.CopyToLocation())

	if i.mechanism == nil {
		ext := filepath.Ext(i.options.Filename)
		i.mechanism = mechanism.NewMechanismByFileExt(ext)
	}
}

func (i ingest) BringOnline(query func(string), username string, password string) {
	i.options.Username = username
	i.options.Password = password
	i.mechanism.BringOnline(i.uri.GetDbNameAsIdentifier(), i.containerId, query, i.options)
	i.setDefaultDatabase(username)
}

func (i ingest) setDefaultDatabase(username string) {
	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		username,
		i.uri.GetDbNameAsNonIdentifier())
	i.query(alterDefaultDb)
}

func (i ingest) IsValidFileExtension() bool {
	for _, m := range mechanism.FileTypes() {
		if m == i.uri.FileExtension() {
			return true
		}
	}
	for _, e := range extract.FileTypes() {
		if e == i.uri.FileExtension() {
			return true
		}
	}
	return false
}

func (i ingest) SourceFileExists() bool {
	return i.location.Exists()
}

func (i ingest) UserProvidedFileExt() string {
	return i.uri.FileExtension()
}

func (i ingest) ValidSchemes() []string {
	return i.location.ValidSchemes()
}

func (i ingest) ValidFileExtensions() []string {
	extensions := []string{}

	for _, m := range mechanism.FileTypes() {
		extensions = append(extensions, m)
	}

	for _, e := range extract.FileTypes() {
		extensions = append(extensions, e)
	}

	return extensions
}
