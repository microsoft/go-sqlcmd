package mechanism

type Mechanism interface {
	FileTypes() []string
	CopyToLocation() string
	BringOnline(databaseName string, containerId string, query func(string), options BringOnlineOptions)
	Name() string
}
