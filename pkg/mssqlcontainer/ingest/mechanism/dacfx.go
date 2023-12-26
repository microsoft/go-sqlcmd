package mechanism

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type dacfx struct {
	controller  *container.Controller
	containerId string
}

func (m *dacfx) Initialize(controller *container.Controller) {
	m.controller = controller
}

func (m *dacfx) CopyToLocation() string {
	return "/var/opt/mssql/backup"
}

func (m *dacfx) Name() string {
	return "dacfx"
}

func (m *dacfx) FileTypes() []string {
	return []string{"bacpac", "dacpac"}
}

func (m *dacfx) BringOnline(
	databaseName string,
	containerId string,
	query func(string),
	options BringOnlineOptions,
) {
	m.containerId = containerId
	m.installSqlPackage()
	m.setDefaultDatabaseToMaster(options.Username, query)

	_, stderr, _ := m.RunCommand([]string{
		"/home/mssql/.dotnet/tools/sqlpackage",
		"/Diagnostics:true",
		"/Action:import",
		"/SourceFile:" + m.CopyToLocation() + "/" + options.Filename,
		"/TargetServerName:localhost",
		"/TargetDatabaseName:" + databaseName,
		"/TargetTrustServerCertificate:true",
		"/TargetUser:" + options.Username,
		"/TargetPassword:" + options.Password,
	})

	if len(stderr) == 0 {
		// Remove the source bacpac file
		m.RunCommandAsRoot([]string{"rm", m.CopyToLocation() + "/" + options.Filename})
	}
}

func (m *dacfx) setDefaultDatabaseToMaster(username string, query func(string)) {
	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		username,
		"master")
	query(alterDefaultDb)
}

func (m *dacfx) installSqlPackage() {
	if m.controller == nil {
		panic("controller is nil")
	}

	m.installDotNet()

	// Check if sqlpackage is installed, if not, install it
	_, stderr, _ := m.RunCommand([]string{"/home/mssql/.dotnet/tools/sqlpackage", "/version"})
	if len(stderr) > 0 {
		m.RunCommand([]string{"/opt/dotnet/dotnet", "tool", "install", "-g", "microsoft.sqlpackage"})
	}
}

func (m *dacfx) installDotNet() {
	// Check if dotnet is installed, if not, install it
	_, stderr, _ := m.RunCommand([]string{"/opt/dotnet/dotnet", "--version"})
	if len(stderr) > 0 {
		// Download dotnet-install.sh and run it
		m.RunCommand([]string{"wget", "https://dot.net/v1/dotnet-install.sh", "-O", "/tmp/dotnet-install.sh"})
		m.RunCommand([]string{"chmod", "+x", "/tmp/dotnet-install.sh"})
		m.RunCommand([]string{"/tmp/dotnet-install.sh", "--install-dir", "/opt/dotnet"})

		// The SQL Server container doesn't have a /home/mssql directory (which is ~), this
		// causes all sorts of things to break in the container that expect to create .toolname folders
		m.RunCommandAsRoot([]string{"mkdir", "-p", "/home/mssql"})
		m.RunCommandAsRoot([]string{"chown", "mssql:root", "/home/mssql"})

		// Add dotnet to the path
		m.AddTextLineToFile(
			"export DOTNET_ROOT=/opt/dotnet",
			"/home/mssql/.bashrc",
		)
		m.AddTextLineToFile(
			"export PATH=$PATH:$DOTNET_ROOT:/home/mssql/.dotnet/tools",
			"/home/mssql/.bashrc",
		)
	}
}

func (m *dacfx) AddTextLineToFile(text string, file string) ([]byte, []byte, int) {
	return m.RunCommand([]string{"/bin/bash", "-c", fmt.Sprintf("echo '%v' >> %v", text, file)})
}

func (m *dacfx) RunCommand(s []string) ([]byte, []byte, int) {
	return m.controller.RunCmdInContainer(m.containerId, s, container.ExecOptions{
		Env: []string{"DOTNET_ROOT=/opt/dotnet"},
	})
}

func (m *dacfx) RunCommandAsRoot(s []string) ([]byte, []byte, int) {
	return m.controller.RunCmdInContainer(m.containerId, s, container.ExecOptions{
		User: "root",
	})
}
