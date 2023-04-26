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

func (m *bacpac) BringOnline(databaseName string, containerId string, query func(string), options BringOnlineOptions) {
	m.containerId = containerId
	m.installSqlPackage()
	m.setDefaultDatabaseToMaster(options.Username, query)

	m.RunCommand([]string{
		"./root/.dotnet/tools/sqlpackage",
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

	/*
	   m.controller.DownloadFile(
	   		m.containerId,
	   		"https://aka.ms/sqlpackage-linux",
	   		"/tmp",
	   	)
	*/
	// wget https://dot.net/v1/dotnet-install.sh -O dotnet-install.sh
	// sudo chmod +x ./dotnet-install.sh
	// ./dotnet-install.sh
	// dotnet tool install -g microsoft.sqlpackage

	m.RunCommand([]string{"wget", "https://dot.net/v1/dotnet-install.sh", "-O", "dotnet-install.sh"})
	m.RunCommand([]string{"chmod", "+x", "./dotnet-install.sh"})
	m.RunCommand([]string{"./dotnet-install.sh"})
	m.RunCommand([]string{"/root/.dotnet/dotnet", "tool", "install", "-g", "microsoft.sqlpackage"})
	//m.RunCommand([]string{"echo", `export PATH="$PATH:/root/.dotnet/tools"`, ">", "~/.bash_profile"})
}

func (m *bacpac) RunCommand(s []string) ([]byte, []byte) {
	return m.controller.RunCmdInContainer(m.containerId, s)
}
