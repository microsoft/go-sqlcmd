package mechanism

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type bacpac struct {
	controller  *container.Controller
	containerId string
}

func (m *bacpac) Initialize(controller *container.Controller) {
	m.controller = controller
}

func (m *bacpac) CopyToLocation() string {
	return "/var/opt/mssql/backup"
}

func (m *bacpac) Name() string {
	return "dacfx"
}

func (m *bacpac) FileTypes() []string {
	return []string{"bacpac", "dacpac"}
}

func (m *bacpac) BringOnline(
	databaseName string,
	containerId string,
	query func(string),
	options BringOnlineOptions,
) {
	m.containerId = containerId
	m.installSqlPackage()
	m.setDefaultDatabaseToMaster(options.Username, query)

	m.RunCommand([]string{
		"./.dotnet/tools/sqlpackage",
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

func (m *bacpac) setDefaultDatabaseToMaster(username string, query func(string)) {
	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		username,
		"master")
	query(alterDefaultDb)
}

func (m *bacpac) installSqlPackage() {
	if m.controller == nil {
		panic("controller is nil")
	}

	//BUG(stuartpa): Can this be done in the mssql user, don't think it needs root
	m.RunCommand([]string{"wget", "https://dot.net/v1/dotnet-install.sh", "-O", "/tmp/dotnet-install.sh"})
	m.RunCommand([]string{"chmod", "+x", "/tmp/dotnet-install.sh"})
	m.RunCommand([]string{"/tmp/dotnet-install.sh", "--install-dir", "/opt/dotnet"})

	// The SQL Server container doesn't have a /home/mssql directory (which is ~), this
	// causes all sorts of things to break in the container that expect to create .toolname folders
	m.RunCommandAsRoot([]string{"mkdir", "-p", "/home/mssql"})
	m.RunCommandAsRoot([]string{"chown", "mssql:root", "/home/mssql"})

	m.RunCommand([]string{"touch", "/home/mssql/.bashrc"})
	m.RunCommand([]string{"sed", "-i", "$ a/export DOTNET_ROOT=/opt/dotnet", "/home/mssql/.bashrc"})
	m.RunCommand([]string{"sed", "-i", "$ a/export PATH=$PATH:$DOTNET_ROOT:~/.dotnet/tools", "/home/mssql/.bashrc"})
	m.RunCommand([]string{"/opt/dotnet/dotnet", "tool", "install", "-g", "microsoft.sqlpackage"})
}

func (m *bacpac) RunCommand(s []string) ([]byte, []byte) {
	return m.controller.RunCmdInContainer(m.containerId, s, container.ExecOptions{})
}

func (m *bacpac) RunCommandAsRoot(s []string) ([]byte, []byte) {
	return m.controller.RunCmdInContainer(m.containerId, s, container.ExecOptions{
		User: "root",
		Env:  nil,
	})
}
