package ingest

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/uri"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/extract"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/location"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"
	"strings"
)

func NewIngest(databaseUri string, controller *container.Controller, options IngestOptions) Ingest {
	uri := uri.NewUri(databaseUri)

	return &ingest{
		uri:        uri,
		controller: controller,
		location:   location.NewLocation(uri.IsLocal(), uri.ActualUrl(), controller),
		mechanism:  mechanism.NewMechanism(uri.FileExtension(), options.Mechanism, controller),
	}
}

func ValidFileExtensions() string {
	var extensions []string

	for _, m := range mechanism.FileTypes() {
		extensions = append(extensions, m)
	}

	for _, e := range extract.FileTypes() {
		extensions = append(extensions, e)
	}

	return strings.Join(extensions, ", ")
}
