package location

type Location interface {
	Exists() bool
	IsLocal() bool
	CopyToContainer(containerId string, destFolder string)
	ValidSchemes() []string
}
