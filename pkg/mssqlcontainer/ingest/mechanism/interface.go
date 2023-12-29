package mechanism

import "github.com/microsoft/go-sqlcmd/internal/container"

type Mechanism interface {
	FileTypes() []string
	Initialize(controller *container.Controller)
	CopyToLocation() string
	BringOnline(databaseName string, containerId string, query func(string), options BringOnlineOptions)
	Name() string
}
