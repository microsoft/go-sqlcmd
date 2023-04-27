package ingest

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/databaseurl"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/extract"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/location"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"
	"strings"
)

func NewIngest(databaseUrl string, controller *container.Controller, options IngestOptions) Ingest {
	url := databaseurl.NewDatabaseUrl(databaseUrl)

	return &ingest{
		url:        url,
		controller: controller,
		location:   location.NewLocation(url.IsLocal, url.String(), controller),
		mechanism:  mechanism.NewMechanism(url.FileExtension, options.Mechanism, controller),
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
