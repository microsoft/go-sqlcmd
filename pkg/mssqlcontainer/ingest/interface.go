package ingest

type Ingest interface {
	IsRemoteUrl() bool
	IsValidScheme() bool
	IsValidFileExtension() bool
	IsExtractionNeeded() bool

	CopyToLocation() string
	SourceFileExists() bool
	UserProvidedFileExt() string

	CopyToContainer(containerId string)
	Extract()
	BringOnline(query func(string), username string, password string)

	ValidSchemes() []string
	ValidFileExtensions() []string
}
