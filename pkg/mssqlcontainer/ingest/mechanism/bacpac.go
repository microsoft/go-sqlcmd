package mechanism

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type bacpac struct {
	controller  *container.Controller
	containerId string
}

func (m bacpac) CopyToLocation() string {
	return "/var/opt/mssql/backup"
}

func (m bacpac) Name() string {
	return "dacfx"
}

func (m bacpac) FileTypes() []string {
	return []string{"bacpac", "dacpac"}
}

func (m bacpac) BringOnline(databaseName string, containerId string, query func(string), options BringOnlineOptions) {
	m.containerId = containerId
	m.installSqlPackage()
	m.setDefaultDatabaseToMaster(options.Username, query)

	m.RunCommand([]string{
		"/opt/sqlpackage/sqlpackage",
		"/Diagnostics:true",
		"/Action:import",
		"/SourceFile:" + m.CopyToLocation() + "/" + options.Filename,
		"/TargetServerName:localhost",
		"/TargetDatabaseName:" + databaseName,
		"/TargetTrustServerCertificate:true",
		"/TargetUser:" + options.Username,
		"/TargetPassword:" + options.Password,
	})
}

func (m bacpac) setDefaultDatabaseToMaster(username string, query func(string)) {
	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		username,
		"master")
	query(alterDefaultDb)
}

func (m bacpac) installSqlPackage() {
	m.controller.DownloadFile(
		m.containerId,
		"https://aka.ms/sqlpackage-linux",
		"/tmp",
	)

	m.RunCommand([]string{"apt-get", "update"})
	m.RunCommand([]string{"apt-get", "install", "-y", "unzip"})
	m.RunCommand([]string{"unzip", "/tmp/sqlpackage-linux", "-d", "/opt/sqlpackage"})
	m.RunCommand([]string{"rm", "/tmp/sqlpackage-linux"})
	m.RunCommand([]string{"chmod", "+x", "/opt/sqlpackage/sqlpackage"})
}

func (m bacpac) RunCommand(s []string) ([]byte, []byte) {
	return m.controller.RunCmdInContainer(m.containerId, s)
}
