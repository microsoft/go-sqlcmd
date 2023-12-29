package mechanism

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type script struct {
}

func (m *script) Initialize(controller *container.Controller) {
}

func (m *script) CopyToLocation() string {
	return "/var/opt/mssql/backup"
}

func (m *script) Name() string {
	return "script"
}

func (m *script) FileTypes() []string {
	return []string{"sql"}
}

func (m *script) BringOnline(databaseName string, _ string, query func(string), options BringOnlineOptions) {
	if options.Filename == "" {
		panic("Filename is required for restore")
	}

	query(options.Filename)
}
