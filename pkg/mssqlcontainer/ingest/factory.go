package ingest

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/databaseurl"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/extract"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/location"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"
	"strings"
)

func NewIngest(databaseUri string, controller *container.Controller, options IngestOptions) Ingest {
	databaseUrl := databaseurl.NewDatabaseUrl(databaseUri)

	return &ingest{
		uri:        databaseUrl,
		controller: controller,
		location:   location.NewLocation(databaseUrl.IsLocal, databaseUrl.String(), controller),
		mechanism:  mechanism.NewMechanism(databaseUrl.FileExtension, options.Mechanism, controller),
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
