package location

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"path/filepath"
)

type local struct {
	uri        string
	controller *container.Controller
}

func (l local) Exists() bool {
	return file.Exists(l.uri)
}

func (l local) IsLocal() bool {
	return true
}

func (l local) ValidSchemes() []string {
	return []string{"file"}
}

func (l local) CopyToContainer(containerId string, destFolder string) {
	l.controller.RunCmdInContainer(
		containerId,
		[]string{"mkdir", "-p", destFolder},
		container.ExecOptions{},
	)

	l.controller.CopyFile(
		containerId,
		l.uri,
		destFolder,
	)

	_, filename := filepath.Split(l.uri)

	l.controller.RunCmdInContainer(
		containerId,
		[]string{"chown", "mssql:root", destFolder + "/" + filename},
		container.ExecOptions{User: "root"},
	)

	l.controller.RunCmdInContainer(
		containerId,
		[]string{"chmod", "-o-r-u+rw-g+r", destFolder + "/" + filename},
		container.ExecOptions{User: "root"},
	)
}
