package mechanism

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type attach struct {
	controller  *container.Controller
	containerId string
}

func (m attach) CopyToLocation() string {
	return "/var/opt/mssql/data"
}

func (m attach) Name() string {
	return "attach"
}

func (m attach) FileTypes() []string {
	return []string{"mdf"}
}

func (m attach) BringOnline(databaseName string, containerId string, query func(string), options BringOnlineOptions) {
	text := `SET NOCOUNT ON; `

	m.containerId = containerId
	m.setFilePermissions(m.CopyToLocation() + "/" + options.Filename)
	if options.LdfFilename == "" {
		text += `CREATE DATABASE [%s] ON (FILENAME = '%s/%s') FOR ATTACH;`
		query(fmt.Sprintf(
			text,
			databaseName,
			m.CopyToLocation(),
			options.Filename,
		))
	} else {
		m.setFilePermissions(m.CopyToLocation() + "/" + options.LdfFilename)
		text += `CREATE DATABASE [%s] ON (FILENAME = '%s/%s'), (FILENAME = '%s/%s') FOR ATTACH;`
		query(fmt.Sprintf(
			text,
			databaseName,
			m.CopyToLocation(),
			options.Filename,
			m.CopyToLocation(),
			options.LdfFilename,
		))
	}
}

func (m attach) setFilePermissions(filename string) {
	m.RunCommand([]string{"chown", "mssql:root", filename})
	m.RunCommand([]string{"chmod", "-o-r", filename})
	m.RunCommand([]string{"chmod", "-u+rw", filename})
	m.RunCommand([]string{"chmod", "-g+r", filename})
}

func (m attach) RunCommand(s []string) ([]byte, []byte) {
	return m.controller.RunCmdInContainer(m.containerId, s)
}
